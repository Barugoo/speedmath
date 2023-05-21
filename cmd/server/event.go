package main

type EventLogin struct {
	Username string
	OutputCh chan any
}

type EventLogout struct {
	Username string
}

type EventAnswer struct {
	RoundNum  int64  `json:"round_num"`
	Username  string `json:"username"`
	Answer    int64  `json:"answer"`
	Timestamp int64  `json:"timestamp"`
}

type EventSendProblem struct {
	Expression     string
	ExpectedAnswer int64
	RoundNum       int64
}
