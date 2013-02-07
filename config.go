package main

/* Authentication credentials for the admin/password reset pages */
const AdminPassword = "bitchmittens"
const AdminRealm    = "Puzzlehunt Admin"
const TeamRealm     = "Pirate Login"

/* Database/server configuration */
const MongoHost     = "mongodb://127.0.0.1:27017"
const MongoDatabase = "puzzlehunt"
const ListenAddress = ":4000"

/* Constant strings and whatnot */
const InvalidAnswerText = "invalid: answers will be one word (no spaces)"

var UnlockTree = map[int][]int {
  1: []int{6, 7},
  2: []int{7, 8},
  3: []int{8, 9},
  4: []int{9, 10},
  5: []int{6, 10},
  6: []int{11, 12},
  7: []int{12, 13},
  8: []int{13, 14},
  9: []int{14, 15},
  10: []int{11, 15},
  11: []int{16},
  12: []int{16},
  13: []int{16},
  14: []int{16},
  15: []int{16},
}
