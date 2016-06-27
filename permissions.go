package main

import (
	"regexp"
	"strings"
)

type Permission struct {
	Database   string
	Table      string
	Privileges []string
}

func (p *Permission) parseUserFromGrantLine(grantLine string) {
	re := regexp.MustCompile("GRANT (.*) ON (.*)\\.(.*) TO.*")
	p.Privileges = strings.Split(re.ReplaceAllString(grantLine, "$1"), ", ")
	p.Database = strings.Replace(re.ReplaceAllString(grantLine, "$2"), "`", "", -1)
	p.Table = strings.Replace(re.ReplaceAllString(grantLine, "$3"), "`", "", -1)
}