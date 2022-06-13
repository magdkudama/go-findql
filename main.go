package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	pathPtr := flag.String("path", ".", "the path where to search from")
	depthPtr := flag.Int("depth", 10, "the maximum depth to traverse the tree")
	filterPtr := flag.String("filter", "", "the query filter to use")
	verbosePtr := flag.Bool("verbose", false, "print verbose output")

	flag.Parse()

	pathToSearchFrom := *pathPtr
	if _, err := os.Stat(pathToSearchFrom); os.IsNotExist(err) {
		log.Fatal(err)
	}

	if *verbosePtr {
		log.Println("searching from " + pathToSearchFrom)
	}

	absPathToSearchFrom, err := filepath.Abs(pathToSearchFrom)
	if err != nil {
		log.Fatal(err)
	}

	if *verbosePtr {
		log.Println("absolute path is " + absPathToSearchFrom)
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	if *verbosePtr {
		log.Println("in memory database opened")
	}

	defer db.Close()

	createStmt := `
		CREATE TABLE entries (
			name TEXT,
			size INT,
			path TEXT,
			depth INT,
			regular BOOLEAN,
			directory BOOLEAN,
			uid INT,
			gid INT,
			user_name TEXT,
			group_name TEXT,
			permission_owner TEXT,
			permission_group TEXT,
			permission_other TEXT,
			accessed_at DATETIME,
			created_at DATETIME,
			modified_at DATETIME
		);
	`

	_, err = db.Exec(createStmt)
	if err != nil {
		log.Fatal(err)
	}

	if *verbosePtr {
		log.Println("created database table")
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	initialParts := len(strings.Split(absPathToSearchFrom, string(os.PathSeparator)))

	err = filepath.WalkDir(absPathToSearchFrom,
		func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			info, err := d.Info()
			if err != nil {
				return err
			}

			parts := len(strings.Split(path, string(os.PathSeparator)))
			if parts-initialParts > *depthPtr {
				return nil
			}

			stmt, err := tx.Prepare("INSERT INTO entries VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			if err != nil {
				log.Fatal(err)
			}

			aTime := info.Sys().(*syscall.Stat_t).Atimespec
			access := time.Unix(aTime.Sec, aTime.Nsec)

			cTime := info.Sys().(*syscall.Stat_t).Ctimespec
			creation := time.Unix(cTime.Sec, cTime.Nsec)

			mTime := info.Sys().(*syscall.Stat_t).Mtimespec
			modification := time.Unix(mTime.Sec, mTime.Nsec)

			owner := info.Mode().String()[1:4]
			group := info.Mode().String()[4:7]
			other := info.Mode().String()[7:10]

			uid := info.Sys().(*syscall.Stat_t).Uid
			gid := info.Sys().(*syscall.Stat_t).Gid

			userData, err := user.LookupId(strconv.Itoa(int(uid)))
			userName := ""
			if err == nil {
				userName = userData.Name
			}

			groupData, err := user.LookupGroupId(strconv.Itoa(int(gid)))
			groupName := ""
			if err == nil {
				groupName = groupData.Name
			}

			_, err = stmt.Exec(
				info.Name(), info.Size(),
				path, parts-initialParts,
				d.Type().IsRegular(), d.Type().IsDir(),
				uid, gid, userName, groupName,
				owner, group, other,
				access, creation, modification,
			)
			if err != nil {
				log.Fatal(err)
			}
			stmt.Close()

			return nil
		})

	if err != nil {
		if *verbosePtr {
			log.Println(err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	if *verbosePtr {
		log.Println("entries inserted, entering search process...")
	}

	selectQuery := "SELECT path FROM entries"
	if *filterPtr != "" {
		selectQuery += " WHERE " + *filterPtr
	}

	rows, err := db.Query(selectQuery)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var rPath string
		err = rows.Scan(&rPath)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(rPath)
	}

	if err != nil {
		log.Println(err)
	}
}
