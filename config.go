package main

/* Email configuration */
const MailServer     = "localhost:25"
const Round1Username = "afriendof"
const Round1Name     = "A Friend"
const Round2Username = "staff"
const Round2Name     = "Weathermen Staff"
const EmailHost      = "theweathermen.info"  /* and always to/from this host */

/* Authentication credentials for the admin/password reset pages */
const AdminPassword = "bitchmittens"
const AdminRealm    = "Puzzlehunt Admin"
const ResetPassword = "doabarrelroll"
const ResetRealm    = "Puzzlehunt Password Reset"
const LastPassword  = "doabarrelroll"
var ResetAnswers = map[string]string{ "color": "cyan",
                                      "school": "east hartford high school",
                                      "maiden": "anderson" }

/* Database/server configuration */
const MongoHost     = "mongodb://127.0.0.1:27017"
const MongoDatabase = "puzzlehunt"
const ListenAddress = ":4000"

/* Constant strings */
const InvalidEmailFormat = `
Email must contain one and only one word which is the answer.

Be sure to delete the text which is automatically filled in which is
the response to the previous email.
`

const EmailInitialSubject = "Welcome to The Weathermen"
const EmailInitialBody = `
Welcome to The Weathermen!

I, your friend, would like to personally extend you a warm welcome. I greatly
need your help in fixing the weather machine, as without you I fear that it will
never be fixed!

I will be sending you emails to your inbox of different problems I'm having with
the weather machine. Each of these should have a word or possibly a phrase as
an answer. If you could just email me the answer you find as a response to each
problem's respective email, I'll be sure to let you know whether it worked or
not.

I'll be responding to the same email thread, and if the answer didn't happen to
work out, just keep replying to the email thread with new answers once you have
another. As a slight technical detail, be sure that the only word in the email
is the answer you're submitting. I don't want to get confused as to which word
in the email is the answer that you're submitting!

I believe that by working together we can solve this strange pattern of weather
occuring all over the planet! You should have some initial problems in your
inbox already, so good luck!

- Your Friend
`

const EmailCorrectAnswer = "That was the correct answer"

const EmailCorrectMorePuzzles = `
That was the correct answer.

You'll find some more puzzles in your inbox, keep up the good work!
`

const EmailSecondRound = `
Hi {{ .Team.Name }},

Here is one of the puzzles David Snow used for his password reminder:

    {{ .Puzzle.Url }}

Let us know if you manage to crack it by replying to this email with just your
answer (the same format as before).

Thanks,
{{ .From }}
`

const EmailInitialRound = `
Dear {{ .Team.Name }},

Please help me fix the machine by solving the following problem:

    {{ .Puzzle.Url }}

Provide your solution as a reply to this email. The reply should contain one
word and no other text. If you are successful in solving this problem, I may ask
you to help me with other problems. Thank you in advance for your help.

- {{ .From }}
`

const EmailFirstRound = `
Dear {{ .Team.Name }},

Due to your great work so far, I have decided to provide you with another
problem to solve. The documents for this problem can be found here:

    {{ .Puzzle.Url }}

Please reply to this email with your answer and no other text. Thanks again for
your help in fixing the machine.

{{ if .Puzzle.UnlockedWithMeta }}
This puzzle is not available in the copy room. You will be receiving a paper
copy soon, however.
{{ else }}
If you would like to receive a hard copy of this problem, I've left a few in PH
A18A for you to pick up. Tell my assistant your team name to receive the proper
documents.
{{ end }}

- {{ .From }}
`

const EmailMetapuzzle = `
Dear {{ .Team.Name }},

Thank you for your help so far with solving the problems with the machine. It
turns out that all of the problems you are solving are subparts for a bigger
issue with the machine. Here is a description of the issue:

    {{ .Puzzle.Url }}

Solving the 12 subproblems I gave you before should help you solve this larger
problem.

Keep up the good work!

- {{ .From }}
`

func (p *Puzzle) EmailSubject() string {
  return "Puzzle: " + p.Name
}
