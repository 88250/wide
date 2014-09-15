package notification

const (
	EvtGOPATHNotFound  = iota // 事件：找不到环境变量 $GOPATH
	EvtGOROOTNotFound         // 事件：找不到环境变量 $GOROOT
	EvtGocodeNotFount         // 事件：找不到 gocode
	EvtIDEStubNotFound        // 事件：找不到 IDE stub
)
