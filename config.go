package main

/* Authentication credentials for the admin/password reset pages */
const AdminPassword = "shenanigans"
const AdminRealm = "Puzzlehunt Admin"
const TeamRealm = "Frood Login"

/* Database/server configuration */
const MongoHost = "mongodb://127.0.0.1:27017"
const MongoDatabase = "puzzlehunt"
const ListenAddress = ":4000"

/* Constant strings and whatnot */
const InvalidAnswerText = "invalid: answers will be one word (no spaces)"

/*var UnlockTree = map[int][]int{
	1:  []int{6, 7, 15},
	2:  []int{7, 8, 15},
	3:  []int{8, 9, 15},
	4:  []int{9, 10, 15},
	5:  []int{6, 10, 15},
	6:  []int{11, 12},
	7:  []int{12, 13},
	8:  []int{13, 14},
	9:  []int{14, 15},
	10: []int{11, 15},
	15: []int{16},
}*/

var MetaRequired = 3
var MetaUnlockers = []int{1, 2, 3, 4, 5}
var MiniMetaIndex = 16
var MetaIndex = 17

var VideoLinks = []string {
	"http://www.youtube.com/watch?v=VrH0-V0YoGE",
	"http://www.youtube.com/watch?v=wqDmXbnwF6E",
	"http://www.youtube.com/watch?v=quamoTYFoZI",
	"http://www.youtube.com/watch?v=wzTvdwWMf7c",
	"http://www.youtube.com/watch?v=rwR1DMbSo80",
}
