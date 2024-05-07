GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOFMT = $(GOCMD) fmt
GOVET = $(GOCMD) vet

SRCDIR = cmd
BINDIR = bin
DATADIR = assets

CMDNAME = cmd
DEPLOYNAME = bootstrap
DEPLOYZIP = deployment.zip

NEWSTMPL = news.html
NEWSJSON = news.json
NEWSDB = news.sqlite

build: cmdline

lambda: linux

linux: deployable

all: cmdline-bin deployable

cmdline: cmdline-bin common-assets

deployable: deployable-bin common-assets deployable-zip

cmdline-bin:
	CGO_ENABLED=1 $(GOBUILD) -tags cmdline -o $(BINDIR)/$(BINNAME) ./$(SRCDIR)/...

deployable-bin:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -tags lambda -o $(BINDIR)/$(DEPLOYNAME) ./$(SRCDIR)/...

deployable-zip:
	cd $(BINDIR) && rm -f $(DEPLOYZIP) && zip --must-match $(DEPLOYZIP) $(DEPLOYNAME) $(NEWSDB) $(NEWSTMPL)

common-assets:
	./scripts/mkdb.sh $(DATADIR)/$(NEWSJSON) $(BINDIR)/$(NEWSDB)
	cp $(DATADIR)/$(NEWSTMPL) $(BINDIR)/

clean:
	$(GOCLEAN)
	rm -rf $(BINDIR)

dep:
	$(GOGET) -u ./$(SRCDIR)/...
	$(GOMOD) tidy
	$(GOMOD) verify

fmt:
	$(GOFMT) ./$(SRCDIR)/...

vet:
	$(GOVET) -tags cmdline ./$(SRCDIR)/...
	$(GOVET) -tags lambda ./$(SRCDIR)/...
