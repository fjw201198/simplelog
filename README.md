# simplelog
An simple log with rotate by day witch golang

## Installation
```go
go get -u github.com/fjw201198/simplelog
```

## Sample
```go
package main

import (
	"github.com/fjw201198/simplelog"
	"time"
)

func goooooooo(logger *simplelog.Logger, idx int, ch chan int) {
	logger.Info("[%d] wwwwwwwhello, world! %s", idx, "testxxxx");
	logger.Debug("[%d] debug infoxxxxxxxx", idx);
	logger.Warn("[%d] warn infoxxxxxxxxxxx", idx);
	logger.Info("[%d]wwwwwwwhello, world! %s", idx, "testxxxx");
	logger.Debug("[%d]debug infoxxxxxxxx", idx);
	logger.Warn("[%d]warn infoxxxxxxxxxxx", idx);
	ch <- 1;
}

func main() {
	// simplelog.SetCached(false);
	defer simplelog.ExitHook();
	var logger = simplelog.GetLogger("testlog");
	// logger.SetPrintConsole(true);
	var ch chan int = make(chan int, 100);
	for i := 0; i < 100; i++ {
		go goooooooo(logger, i, ch);
	}
	// wait
	for j := 0; j < 100; j++ {
		<-ch;
	}
	time.Sleep(20000000000); // wait 20 seconds for change system time. for test rotate
	for i := 0; i < 100; i++ {
		go goooooooo(logger, i, ch);
	}
	// wait
	for j := 0; j < 100; j++ {
		<-ch;
	}
}
```

## APIs
+ Simple Log
  - [simplelog.SetCached(cache bool)](#SetCached)
  - [simplelog.GetCached() bool](#GetCached)
  - [simplelog.SetLogDir(logdir string)](#SetLogDir)
  - [simplelog.GetLogDir() string](#GetLogDir)
  - [simplelog.SetFlushInterval(seconds int64)](#SetFlushInterval)
  - [simplelog.GetFlushInterval() int64](#GetFlushInterval)
  - [simplelog.FlushAllLogs()](#FlushAllLogs)
  - [simplelog.ExitHook()](#ExitHook)
  - [simplelog.GetLogger() *Logger](#GetLogger)
  - [(*Logger) SetLevel(level uint32)](#SetLevel)
  - [(*Logger) GetLevel() uint32](#GetLevel)
  - [(*Logger) SetPrintConsole(yes bool)](#SetPrintConsole)
  - [(*Logger) GetPrintConsole() bool](#GetPrintConsole)
  - [(*Logger) SetZip(zipped bool)](#SetZip)
  - [(*Logger) GetZip() bool](#GetZip)
  - [(*Logger) WaitForZipComplete()](#WaitForZipComplete)
  - [(*Logger) Debug(format string, args...interface{})](#Debug)
  - [(*Logger) Info(format string, args...interface{})](#Info)
  - [(*Logger) Warn(format string, args...interface{})](#Warn)
  - [(*Logger) Err(format string, args...interface{})](#Err)
  - [(*Logger) Alert(format string, args...interface{})](#Alert)
  - [(*Logger) Fatal(format string, args...interface{})](#Fatal)
 
 1. <a name="SetCached">SetCached</a>
 ```go
 func SetCached(cache bool)
 ```
 set whether use memory cache. this configure has global scope.
 
 2. <a name="GetCached">GetCached</a>
 ```go
 func GetCached() bool
 ```
 get current cache setting.
 
 3. <a name="SetLogDir">SetLogDir</a>
 ```go
 func SetCached(cache bool)
 ```
 set whether use memory cache. this configure has global scope.
 
 4. <a name="GetLogDir">GetLogDir</name>
 ```go
 func GetLogDir() string
 ```
 get current log dir.
 
 5. <a name="SetFlushInterval">SetFlushInterval</a>
 ```go
 func SetFlushInterval(seconds int64)
 ```
 set flush log interval if cache opened. Also, this variable has global scope.
 
 6. <a name="GetFlushInterval">GetFlushInterval</a>
 ```go
 func GetFlushInterval() int64
 ```
 Get current flush log interval.
 
 7. <a name="FlushAllLogs">FlushAllLogs</a>
 ```go
 func FlushAllLogs()
 ```
 Flush logs manually. call this function will flush all log instance.
 we recommends add following call in the main function:
 ```go
 defer simplelog.FlushAllLogs();
 ```
 if you opened the zip flag, you'd better use ExitHook() with defer, rather than this function.

8. <a name="ExitHook">ExitHook</a>
```go
func ExitHook()
```
flush all caches and Wait all zip work complete. use this function better than use FLushAllLogs() if
you opened the zip flag.

9. <a name="GetLogger">GetLogger</a>
```go
func GetLogger(name string) *Logger
```
Get or create logger instance. if the logger named "name" already exist, we will just return it, otherwise,
we'll create a new logger instance and return it.
return the logger instance witch named by "name".

10. <a name="SetLevel">SetLevel</a>
```go
func (self *Logger) SetLevel(level uint32)
```
set log level, default is `INFO`. the level is one of DEBUG, INFO, WARN, ERR, ALERT and FATAL.

11. <a name="GetLevel">GetLevel</a>
```go
func (self *Logger) GetLevel() uint32
```
get current log level.

12. <a name="SetPrintConsole">SetPrintConsole</a>
```go
func (self *Logger) SetPrintConsole(yesNo bool) 
```
set whether append the log to the console. default is `false`.

13. <a name="GetPrintConsole">GetPrintConsole</a>
```go
func (self *Logger) GetPrintConsole() bool
```
get whether append the log to the console.

14. <a name="SetZip">SetZip</a>
```go
func (self *Logger) SetZip(yesNo bool) {
```
set whether zip the log. NOTICE that we'll create a new routine to process the zip work,
you should wait for it complete before the program closed.

15. <a name="GetZip">GetZip</a>
```go
func (self *Logger) GetZip() bool 
```
get whether zip the log flag.

16. <a name="WaitForZipComplete">WaitForZipComplete</a>
```go
func (self *Logger) WaitForZipComplete()
```
wait for zip complete before the program closed.

17. <a name="Debug">Debug</a>
```go
func (self *Logger) Debug(f string, args...interface{})
```
write debug log, only debug level print it.

18. <a name="Info">Info</a>
```go
func (self *Logger) Info(f string, args...interface{})
```
write info log, only INFO and DEBUG level print it.

19. <a name="Warn">Warn</a>
```go
func (self *Logger) Warn(f string, args...interface{})
```
write warn log, only INFO and DEBUG, WARN level print it.

20. <a name="Err">Err</a>
```go
func (self *Logger) Err(f string, args...interface{})
```
write error log, only WARN, ERR, INFO and DEBUG level print it.

21. <a name="Alert">Alert</a>
```go
func (self *Logger) Alert(f string, args...interface{})
```
write Alert log, only Fatal level NOT print it.

22. <a name="Fatal">Fatal</a>
```go
func (self *Logger) Fatal(f string, args...interface{})
```
write Fatal log, all log level print it.
