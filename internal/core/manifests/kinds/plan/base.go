package plan

type StageMode string

const (
	Strict   StageMode = "strict"
	Parallel StageMode = "parallel"
)

func (s StageMode) String() string {
	return string(s)
}
