package initializer

const (
	Logger = iota
	Conf
	Etcd
	Grpc
	Http
)

var (
	Sequence = 5
	initMap  = map[int]func(*Global){
		Logger: InitializeLogger,
		Conf:   InitializeConfig,
		Etcd:   InitializeEtcd,
	}
)

// InitialSequence 执行标识顺序
func InitialSequence() *Global {
	glb := new(Global)
	for i := 0; i < Sequence; i++ {
		f, ok := initMap[i]
		if ok {
			f(glb)
		}
	}

	glb.Logger.Info("Sequence initialized")
	return glb
}
