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

	s := grpc.NewServer()

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
