#!/usr/bin/env bash

# exit if any command fails
set -e

JSON="$1"
DB="$2"

[ "$DB" = "" ] && DB="news.sqlite"

if [ ! -f "$JSON" ]; then
	echo "invalid json file: [$JSON]"
	exit 1
fi

[ -f "$DB" ] && rm -f "$DB"

cat <<EOF | sqlite3 "$DB"

CREATE TABLE microfilm (
	state  TEXT,
	abbrev TEXT,
	city   TEXT,
	title  TEXT,
	begin  NUMBER,
	end    NUMBER,
	callno TEXT
);

INSERT INTO microfilm
	SELECT
		json_extract(value, '$.state'),
		json_extract(value, '$.abbrev'),
		json_extract(value, '$.city'),
		json_extract(value, '$.title'),
		json_extract(value, '$.begin'),
		json_extract(value, '$.end'),
		json_extract(value, '$.callno')
	FROM
		json_each(readfile('${JSON}'));

EOF

exit 0
