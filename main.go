package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"mime"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 设置一个端口号，用于监听
	port := "6789"

	// 监听端口
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	fmt.Println("===========正在监听 6789 端口===========")

	// 建立无限循环，等待客户端请求
	for {
		// 接受客户端连接请求
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
			continue
		}

		// 创建新的goroutine处理客户端请求
		go handleRequest(conn)
	}
}

// 处理客户端请求
func handleRequest(conn net.Conn) {

	// 从连接中读取请求
	reader := bufio.NewReader(conn)

	// 读取请求行
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	// 控制台打印请求行
	fmt.Println("==========请求行==========")
	fmt.Println(requestLine)

	// 从连接中读取头部行
	// 头部行和实体行之间有个空行，借此读取头部行
	var headerLines []string
	for {
		headerLine, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}
		// 意味着读取了一个空行
		if len(headerLine) == 2 && headerLine == "\r\n" {
			break
		}
		// 每读取一行 就添加到 headerLines中
		headerLines = append(headerLines, headerLine)
	}
	// 控制台打印头部行
	fmt.Println("==========头部行==========")
	fmt.Println(headerLines)

	// 解析请求行并获取文件路径 遇到空格就截取一个
	requestParts := strings.Split(requestLine, " ")
	if len(requestParts) != 3 {
		fmt.Println("Invalid request line: ", requestLine)
		return
	}

	method := requestParts[0]
	path := requestParts[1]
	httpVersion := requestParts[2]

	fmt.Println("==========解析请求行==========")

	fmt.Println("request method: ", method)
	fmt.Println("incomplete file path: ", path)
	fmt.Println("httpVersion", httpVersion)

	// favicon.ico是浏览器在加载网页时默认请求的网站图标, 以下代码是为了避免这个问题 ,但是每次都请求问题没有解决, 当然也有可能是我的浏览器设置的问题吧
	// 如果请求的文件路径为 /favicon.ico, 直接返回空响应即可

	if path == "/favicon.ico" {
		conn.Close()
		return
	}

	// 构建完整路径, 其实也是相对路径(相对于main.go文件)
	fullPath := filepath.Join(".", path)

	// 打开文件
	file, err := os.Open(fullPath)
	if err != nil {
		fmt.Println("Error opening file:", err.Error())
		return
	}

	// 读取文件结束后关闭
	defer file.Close()

	// 读取文件内容
	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err.Error())
		return
	}
	// 获取 MIME 类型 但是一些txt类型在go的mime库中没有设置，需要我们自己判断, 这里不在判断
	// 所以如果我们请求文件为txt的话， 浏览器会下载文件到本地
	mimeType := mime.TypeByExtension(filepath.Ext(fullPath))

	// 如果没有获取到MIME类型, 默认设置为"application/octet-stream"
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// 打开注释即可加上txt的判断， 也可以单独写一个判断的函数
	// // 简单判断一个txt文件类型, 因为mime库里没有设置txt类型
	// if strings.HasSuffix(fullPath, ".txt") {
	// 	mimeType = "text/plain"
	// }

	// 发送响应头
	responseHeader := "HTTP/1.1 200 OK\r\n"
	responseHeader += fmt.Sprintf("Content-Type: %s\r\n", mimeType)
	responseHeader += fmt.Sprintf("Content-Length: %d\r\n", len(fileContents))
	responseHeader += "\r\n"
	conn.Write([]byte(responseHeader))

	// 发送文件内容
	conn.Write(fileContents)

	// 关闭连接
	conn.Close()

	// 一次请求响应结束

	fmt.Println("==========一次请求响应结束, 等待下一次请求==========")
}
