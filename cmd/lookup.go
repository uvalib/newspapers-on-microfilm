package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// database connection
var db *sql.DB

// prepared statements
var stmtState *sql.Stmt
var stmtRange *sql.Stmt
var stmtStateRange *sql.Stmt

// html template
var tmpl *template.Template

// structs for building html response from template
type Entry struct {
	State  string
	City   string
	Title  string
	Begin  int
	End    int
	CallNo string
}

type StateEntry struct {
	State   string
	Entries []Entry
}

type NewsResults struct {
	Header       string
	StateEntries []StateEntry
}

// request structure passed by caller
type lookupRequest struct {
	choice string
	state  string
	year   string
	begin  string
	end    string
}

// response structure returned to caller
type lookupResponse struct {
	results NewsResults
	err     error
}

func lookup(req lookupRequest) lookupResponse {
	var header string
	var rows *sql.Rows
	var err error

	// enforce upper case for backwards compatibility, and ease of comparisons
	req.state = strings.ToUpper(req.state)

	switch req.choice {
	case "1":
		if req.year == "" {
			return lookupResponse{err: errors.New("choice 1 requires a year")}
		}

		header = req.year
		rows, err = stmtRange.Query(req.year, req.year)

	case "2":
		if req.state == "" {
			return lookupResponse{err: errors.New("choice 2 requires a state")}
		}

		header = req.state
		rows, err = stmtState.Query(req.state)

	case "3":
		if req.state == "" || req.begin == "" || req.end == "" {
			return lookupResponse{err: errors.New("choice 3 requires a state, beginning year, and ending year")}
		}

		header = fmt.Sprintf("%s %s - %s", req.state, req.begin, req.end)

		if req.state == "ALL STATES" {
			rows, err = stmtRange.Query(req.begin, req.end)
		} else {
			rows, err = stmtStateRange.Query(req.state, req.begin, req.end)
		}

	default:
		return lookupResponse{err: errors.New("invalid choice")}
	}

	if err != nil {
		return lookupResponse{err: fmt.Errorf("[SQL] failed to execute query: [%s]", err.Error())}
	}

	defer rows.Close()

	// collect results in a per-state map, along with a list of all states seen.
	// the query results are ordered by state, and the list of all states
	// will retain that order when building the final result set below.

	var states []string
	stateMap := make(map[string][]Entry)

	curState := ""
	for rows.Next() {
		var e Entry

		err = rows.Scan(&e.State, &e.City, &e.Title, &e.Begin, &e.End, &e.CallNo)
		if err != nil {
			return lookupResponse{err: fmt.Errorf("[SQL] failed to scan row: [%s]", err.Error())}
		}

		// this approach for collecting states works because the results are ordered by state
		if e.State != curState {
			curState = e.State
			states = append(states, curState)
		}

		stateMap[curState] = append(stateMap[curState], e)
	}

	// check for errors

	err = rows.Err()
	if err != nil {
		return lookupResponse{err: fmt.Errorf("[SQL] select failed: [%s]", err.Error())}
	}

	// build results from states map above

	results := NewsResults{Header: strings.ToUpper(header)}

	// assemble a list of entries per state, in the order returned by the query
	for _, state := range states {
		results.StateEntries = append(results.StateEntries, StateEntry{State: state, Entries: stateMap[state]})
	}

	return lookupResponse{results: results}
}

func (r lookupResponse) toText() (string, error) {
	// determine max length of variable-length city/title fields for aligning columns

	maxCity := 0
	maxTitle := 0

	for _, s := range r.results.StateEntries {
		for _, e := range s.Entries {
			if len(e.City) > maxCity {
				maxCity = len(e.City)
			}
			if len(e.Title) > maxTitle {
				maxTitle = len(e.Title)
			}
		}
	}

	// build wall of text

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%s\n", r.results.Header)
	fmt.Fprintf(&buf, "\n")

	for _, s := range r.results.StateEntries {
		if len(r.results.StateEntries) != 1 || strings.ToUpper(s.State) != strings.ToUpper(r.results.Header) {
			fmt.Fprintf(&buf, "%s\n", s.State)
			fmt.Fprintf(&buf, "\n")
		}

		for _, e := range s.Entries {
			fmt.Fprintf(&buf, "%s%s    ", e.City, strings.Repeat(" ", maxCity-len(e.City)))
			fmt.Fprintf(&buf, "%s%s    ", e.Title, strings.Repeat(" ", maxTitle-len(e.Title)))
			fmt.Fprintf(&buf, "%d    %d    %s\n", e.Begin, e.End, e.CallNo)
		}

		fmt.Fprintf(&buf, "\n")
	}

	return buf.String(), nil
}

func (r lookupResponse) toHTML() (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r.results); err != nil {
		return "", fmt.Errorf("[HTML] failed to execute template: [%s]", err.Error())
	}
	return buf.String(), nil
}

func init() {
	var err error

	// open a persistent connection to the database
	if db, err = sql.Open("sqlite3", "news.sqlite"); err != nil {
		log.Fatalf("[SQL] failed to open database: [%s]", err.Error())
	}

	// assemble prepared statements for each type of query that might be executed from the following query fragments:
	selectClause := "SELECT state, city, title, begin, end, callno FROM microfilm"
	whereStateClause := "UPPER(?) IN (UPPER(state), UPPER(abbrev))"
	whereRangeClause := "? <= end and ? >= begin"
	orderClause := "ORDER BY state ASC, city ASC, title ASC, begin ASC, end ASC, callno ASC"

	// prepare query by state
	if stmtState, err = db.Prepare(fmt.Sprintf("%s WHERE (%s) %s", selectClause, whereStateClause, orderClause)); err != nil {
		log.Fatalf("[SQL] failed to prepare state statement: [%s]", err.Error())
	}

	// prepare query by year range
	if stmtRange, err = db.Prepare(fmt.Sprintf("%s WHERE (%s) %s", selectClause, whereRangeClause, orderClause)); err != nil {
		log.Fatalf("[SQL] failed to prepare range statement: [%s]", err.Error())
	}

	// prepare query by state and year range
	if stmtStateRange, err = db.Prepare(fmt.Sprintf("%s WHERE (%s) AND (%s) %s", selectClause, whereStateClause, whereRangeClause, orderClause)); err != nil {
		log.Fatalf("[SQL] failed to prepare state range statement: [%s]", err.Error())
	}

	// load the template used for html responses
	if tmpl, err = template.ParseFiles("news.html"); err != nil {
		log.Fatalf("[TEMPLATE] failed to parse html template: [%s]", err.Error())
	}
}
