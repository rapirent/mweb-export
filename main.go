package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
    "bufio"
    "path/filepath"
    "strings"
)

type Category struct {
	PID         uint64
	UUID        uint64
	Name        string
	SubCategory []*Category
	Article     []*Article
}

type Article struct {
	RID   uint64
	AID   uint64
	Name  string
	Media []string
}

func (a *Article) update(root string, target string, catMap map[uint64]*Category) {
    name := ""
    var ferr, err error
	if file, ferr := os.Open(filepath.Join(root, fmt.Sprintf("%d.md", a.AID)));  ferr == nil {
        defer file.Close()
        scanner := bufio.NewScanner(file)
        scanner.Scan()
        name = scanner.Text()[2:]
        name = strings.Replace(name, "/", "-", -1)
	} else {
		log.Print(ferr)
    }


    if name != "" {
        if _, err := os.Stat(filepath.Join(target, catMap[a.RID].Name, fmt.Sprintf("%s.md", name))); os.IsNotExist(err) {
            err = os.Rename(filepath.Join(root, fmt.Sprintf("%d.md", a.AID)), filepath.Join(target, catMap[a.RID].Name, fmt.Sprintf("%s.md", name)))
        }
    } else {
        err = nil
    }

    if ferr != nil || err != nil || name == "" {
        if err = os.Rename(filepath.Join(root, fmt.Sprintf("%d.md", a.AID)), filepath.Join(target, catMap[a.RID].Name, fmt.Sprintf("%d.md", a.AID))) ; err != nil {
            log.Printf("can't change filepath to %s \n due to %v \n",
                filepath.Join(target, catMap[a.RID].Name, fmt.Sprintf("%d.md", a.AID)), err)
            return
        }
    }

    if _, err := os.Stat(filepath.Join(root, "media", fmt.Sprintf("%d", a.AID))); os.IsNotExist(err) {
        return
    }
    var files []string
    err = filepath.Walk(filepath.Join(root, "media", fmt.Sprintf("%d", a.AID)), func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            files = append(files, path)
        }
        return nil
    })
    if err != nil {
        log.Fatalf("read directory error due to %v", err)
    }
    for _, file := range files {
        if err := os.MkdirAll(filepath.Join(target, catMap[a.RID].Name, "media", fmt.Sprintf("%d", a.AID)), os.ModePerm); err != nil {
            log.Fatalf("Error while creating media directory under %s/%d", catMap[a.RID].Name, a.AID)
        }
        if err := os.Rename(file, filepath.Join(target, catMap[a.RID].Name, "media", fmt.Sprintf("%d", a.AID), filepath.Base(file))); err != nil {
            log.Fatalf("can't change path %s to %s", file, filepath.Join(target, catMap[a.RID].Name, "media", fmt.Sprintf("%d", a.AID), filepath.Base(file)))
        }
    }
}

func tree(cat *Category, deep int, buff *bytes.Buffer) {
	space := ""
	for i := 0; i < deep; i++ {
		space = space + "  "
	}
	buff.WriteString(fmt.Sprintf("%s- %s\n", space, cat.Name))
	for _, article := range cat.Article {
		buff.WriteString(fmt.Sprintf("%s  - [%s](./docs/%d.md)\n", space, article.Name, article.AID))
	}
	for _, category := range cat.SubCategory {
		tree(category, deep+1, buff)
	}
}

func categories(db *sql.DB) ([]*Category, error) {
	row, err := db.Query("select pid, uuid, name from cat")
	if err != nil {
		return nil, err
	}
	var temp []*Category
	for row.Next() {
		var (
			pid  uint64
			uuid uint64
			name string
		)
		if err := row.Scan(&pid, &uuid, &name); err == nil {
			temp = append(temp, &Category{
				PID:  pid,
				UUID: uuid,
				Name: name,
			})
		}
	}
	return temp, nil
}

func article(db *sql.DB) ([]*Article, error) {
	row, err := db.Query("select rid, aid from cat_article")
	if err != nil {
		return nil, err
	}
	var temp []*Article
	for row.Next() {
		var (
			rid uint64
			aid uint64
		)
		if err := row.Scan(&rid, &aid); err == nil {
			temp = append(temp, &Article{
				RID: rid,
				AID: aid,
			})
		}
	}
	return temp, nil
}

func makeCategoryTree(root *Category, input []*Category) []*Category {
	var otherNode []*Category
	for _, category := range input {
		if category.PID == root.UUID {
			root.SubCategory = append(root.SubCategory, category)
		} else {
			otherNode = append(otherNode, category)
		}
	}
	for _, category := range root.SubCategory {
		otherNode = makeCategoryTree(category, otherNode)
	}
	return otherNode
}

func main() {
	pwd, _ := os.Getwd()

	lib := flag.String("path", pwd, "path to MWebLibrary")
	target := flag.String("target", pwd, "export README.md directory")
	help := flag.Bool("help", false, "show usage")

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *lib == "" {
		log.Fatalf("You must set MWebLibrary path")
	}

	sqlPath := filepath.Join(*lib, "mainlib.db")
	db, dErr := sql.Open("sqlite3", sqlPath)
	if dErr != nil {
		log.Fatalf("Open database  fail: %v", dErr)
	}

	cat, cErr := categories(db)
	if cErr != nil {
		log.Fatalf("Read categories fail: %v", cErr)
	}

	art, aErr := article(db)
	if aErr != nil {
		log.Fatalf("Read article fail: %v", dErr)
	}

	catMap := map[uint64]*Category{}

	for _, category := range cat {
		catMap[category.UUID] = category

        if err := os.MkdirAll(filepath.Join(*target, category.Name), os.ModePerm); err != nil {
            log.Fatalf("Error while creating directory: %s", category.Name)
        }
        if err := os.MkdirAll(filepath.Join(*target, category.Name, "media"), os.ModePerm); err != nil {
            log.Fatalf("Error while creating media directory under %s", category.Name)
        }
	}
	for _, article := range art {
		article.update(filepath.Join(*lib, "docs"), *target, catMap)
	}
}
