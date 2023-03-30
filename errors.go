package bridgemq

type Err struct {
	Msg  string
	Code int
}

func (e Err) String() string {
	return e.Msg
}

func (e Err) Error() string {
	return e.Msg
}

var (
	ErrInvalidBroker = Err{Code: 10000, Msg: "invalid broker, borker can not be nil"}
)
