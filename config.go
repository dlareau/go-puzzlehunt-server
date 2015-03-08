package main

/* Authentication credentials for the admin/password reset pages */
const AdminPassword = "secure"
const AdminRealm    = "Puzzlehunt Admin"
const TeamRealm     = "Hunter Login"

/* Database/server configuration */
const MongoHost     = "mongodb://127.0.0.1:27017"
const MongoDatabase = "puzzlehunt"
const ListenAddress = ":80"

/* Constant strings and whatnot */
const InvalidAnswerText = "invalid: answers will be one word (no spaces)"

var UnlockTree = map[int][]int {
  1: []int{6, 10},
  2: []int{7, 6},
  3: []int{8, 7},
  4: []int{9, 8},
  5: []int{10, 9},
  6: []int{13, 12},
  7: []int{14, 13},
  8: []int{15, 14},
  9: []int{11, 15},
  10: []int{12, 11},
  11: []int{16},
  12: []int{16},
  13: []int{16},
  14: []int{16},
  15: []int{16},
  17: []int{18, 19},
  18: []int{20, 21},
  19: []int{20, 21, 22},
  20: []int{22, 23},
  21: []int{22, 23},
  22: []int{23},
}

var MetaRequired = 3
var MetaMinimum = 8
var MetaIndex = 16
