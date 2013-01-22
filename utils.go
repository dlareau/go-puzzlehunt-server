package main

import "strings"
import "unicode"
import "regexp"

var whitespace = regexp.MustCompile(`\s+`)

func Parameterize(s string) string {
  s = strings.Map(func (r rune) rune {
    switch {
      case 'A' <= r && r <= 'Z': return unicode.ToLower(r)
      case 'a' <= r && r <= 'z': return r
      case '0' <= r && r <= '9': return r
    }
    return ' '
  }, s)
  return whitespace.ReplaceAllString(s, "-")
}
