package main

import (
	"regexp"
	"sort"
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
	sort.Strings(p.Privileges)
	p.Database = strings.Replace(re.ReplaceAllString(grantLine, "$2"), "`", "", -1)
	p.Table = strings.Replace(re.ReplaceAllString(grantLine, "$3"), "`", "", -1)
}

func (p *Permission) compare(pr Permission) bool {
	if p.Database == pr.Database && p.Table == pr.Table && len(p.Privileges) == len(pr.Privileges) {
		for _, priv1 := range p.Privileges {
			var found bool
			for _, priv2 := range pr.Privileges {
				if priv1 == priv2 {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	} else {
		return false
	}
}
