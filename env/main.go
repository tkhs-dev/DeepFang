package main

import (
	"fmt"
	"golib/server"
	grpc_gen "golib/server/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"

	ort "github.com/yalue/onnxruntime_go"
	"google.golang.org/grpc"
)

func main() {
	if runtime.GOOS == "windows" {
		ort.SetSharedLibraryPath("thirdparty/onnxruntime.dll")
	} else {
		ort.SetSharedLibraryPath("thirdparty/libonnxruntime.so")
	}
	err := ort.InitializeEnvironment()
	if err != nil {
		println("Failed to initialize ONNX Runtime environment:", err.Error())
		return
	}

	port := 8080
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer(grpc.MaxRecvMsgSize(8 * 1024 * 1024))

	grpc_gen.RegisterEnvServiceServer(s, server.NewServer())

	go func() {
		log.Printf("start gRPC server port: %v", port)
		s.Serve(listener)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("stopping gRPC server...")
	s.GracefulStop()
}

//func main() {
//	if runtime.GOOS == "windows" {
//		ort.SetSharedLibraryPath("thirdparty/onnxruntime.dll")
//	} else {
//		ort.SetSharedLibraryPath("thirdparty/libonnxruntime.so")
//	}
//	err := ort.InitializeEnvironment()
//	if err != nil {
//		println("Failed to initialize ONNX Runtime environment:", err.Error())
//		return
//	}
//	bytes, err := ioutil.ReadFile("../output.onnx")
//	if err != nil {
//		panic(err)
//	}
//	model, err := neuralnet.NewOnnxModel(bytes, 1, 6, 6)
//	if err != nil {
//		println("Failed to load model:", err.Error())
//		return
//	}
//	defer model.Close()
//	data := neuralnet.GetInputTensor(rule.NewBoardEditor(6, 6).PlaceFoods(rule.Point{5, 5}).PlaceSnake([]rule.Point{{1, 1}, {1, 2}, {1, 3}}, 100).Build())
//	predict, f, err := model.Predict(data)
//	if err != nil {
//		println("Model prediction failed:", err.Error())
//		return
//	}
//	fmt.Println(data.CNNInput)
//	fmt.Println(data.FeatureInput)
//	fmt.Println(predict, f)
//}
