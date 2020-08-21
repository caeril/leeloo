package boojum

type ErrorHandler func(error)

type Boojum interface {
	Set([]byte) error
	Get() ([]byte, error)
	Cleanup() error
}

func Init(f ErrorHandler) Boojum {

	return CreateNaiveBoojum(f)
}
