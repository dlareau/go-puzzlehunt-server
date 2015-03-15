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

var MetaRequired = 3
var MetaMinimum = 10

var UnlockTree = map[int][]int {
  1: []int{6, 10},
  2: []int{9, 10},
  3: []int{8, 9},
  4: []int{7, 8},
  5: []int{6, 7},
  6: []int{11, 15},
  7: []int{14, 15},
  8: []int{13, 14},
  9: []int{12, 13},
  10: []int{11, 12},
  11: []int{16},
  12: []int{16},
  13: []int{16},
  14: []int{16},
  15: []int{16},
}

var MetaIndex = 16
