package simplelog

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"
	"unsafe"
)

/* define log levels */
const (
	FATAL 	uint32 = 0
	ALERT 	uint32 = 1
	ERR		uint32 = 2
	WARN    uint32 = 3
	INFO    uint32 = 4
	DEBUG   uint32 = 5
)

const (
	BUF_SIZE 	int = 500 			// cache log message 200 records.
	BUF_EXPIRES int64 = 1 			// cache log message 1 second.
	DFT_LOG_DIR string = "./logs"	// default log directory
)

type logRecord struct {
	logtime time.Time;
	data    string;
}

type Logger struct {
	offset 	 	uint32;
	flushTime 	time.Time;
	name   		string;
	level       uint32;
	logfile    *os.File;
	logObj     *log.Logger;
	lock        sync.Mutex;
	console     bool;
	zipped      bool;
	zipping     bool;
	zipChan     chan int;
	/* cache  	 []*logRecord; */
	cache        []string;
};

type logManager struct {
	logdir 			string;
	cached 			bool;
	flushInterval   int64;
	logs map[string]*Logger;
};

var logMgr = &logManager{DFT_LOG_DIR, true, BUF_EXPIRES,make(map[string]*Logger)};

// public methods ....
/**
 * void SetCached(bool cache)
 * set whether use memory cache. this configure has global scope.
 * @param cache if cache is true, the cache switcher will be used.
 */
func SetCached(cache bool) {
	logMgr.cached = cache;
}

/**
 * bool GetCache()
 * get current cache setting.
 * @return if global cache is open, then return true, else return false.
 */
func GetCached() bool {
	return logMgr.cached;
}

/**
 * void SetLogDir(string dir);
 * set global log stored directory, default is `./logs'. Notice that, you must call it before you create Logger.
 * @param dir the directory witch will be used.
 */
func SetLogDir(logdir string) {
	logMgr.logdir = logdir;
}

/**
 * string GetLogDir();
 * get current log dir.
 * @return return the current log directory.
 */
func GetLogDir() string {
	return logMgr.logdir;
}

/**
 * void SetFlushInterval(int64 sec);
 * set flush log interval if cache opened. Also, this variable has global scope.
 * @param sec set the flush log interval to `sec' seconds.
 */
func SetFlushInterval(seconds int64) {
	logMgr.flushInterval = seconds;
}

/**
 * int64 GetFlushInterval();
 * Get current flush log interval
 * @return return the global flush log interval.
 */
func GetFlushInterval() int64 {
	return logMgr.flushInterval;
}

/**
 * void FlushALlLogs();
 * Flush logs manually. call this function will flush all log instance.
 * we recommends add following call in the main function:
 * ```go
 * 	defer simplelog.FlushAllLogs();
 * ```
 * if you opened the zip flag, you'd better use ExitHook() with defer, rather than this function.
 */
func FlushAllLogs() {
	curTime := time.Now();
	for _, logger := range logMgr.logs {
		logger.flushCache();
		logger.flushTime = curTime;
	}
}

/**
 * void ExitHook();
 * flush all caches and Wait all zip work complete. use this function better than use FLushAllLogs() if
 * you opened the zip flag.
 */
func ExitHook() {
	for _, logger := range logMgr.logs {
		logger.WaitForZipComplete();
	}
	FlushAllLogs();
}

/**
 * Logger* GetLogger(string name);
 * Get or create logger instance. if the logger named `name' already exist, we will just return it, otherwise,
 * we'll create a new logger instance and return it.
 * return the logger instance witch named by `name'.
 */
func GetLogger(name string) *Logger {
	logger, ok := logMgr.logs[name];
	if !ok {
		logger = newLogger(name);
	}
	logMgr.logs[name] = logger;
	return logger;
}

/**
 * void Logger::SetLevel(uint32 level);
 * set log level, default is `INFO`.
 * @param level current log level.
 */
func (self *Logger) SetLevel(level uint32) {
	self.level = level;
}

/**
 * uint32 Logger::GetLevel();
 * get current log level.
 * @return return current log level.
 */
func (self *Logger) GetLevel() uint32 {
	return self.level;
}

/**
 * void Logger::SetPrintConsole(bool yesNo);
 * set whether append the log to the console.
 * @param yesNo true for yes, false for no
 */
func (self *Logger) SetPrintConsole(yesNo bool) {
	self.console = yesNo;
}

/**
 * bool Logger::GetPrintConsole();
 * get whether append the log to the console.
 * @return if yes return true, no return false
 */
func (self *Logger) GetPrintConsole() bool {
	return self.console;
}

/**
 * void Logger::SetZip(bool yesNo);
 * set whether zip the log. NOTICE that we'll create a new routine to process the zip work,
 * you should wait for it complete before the program closed.
 * @param yesNo true for yes, false for no
 */
func (self *Logger) SetZip(yesNo bool) {
	self.zipped = yesNo;
}

/**
 * bool Logger::GetZip();
 * get whether zip the log flag.
 * @return return where opened the zip switcher, true for yes, false for no.
 */
func (self *Logger) GetZip() bool {
	return self.zipped;
}

/**
 * void Logger::WaitForZipComplete();
 * wait for zip complete before the program closed.
 */
func (self *Logger) WaitForZipComplete() {
	if !self.zipped || !self.zipping { return; }
	for {
		select {
		case <- self.zipChan:
			return;
		}
	}
}

/**
 * void Logger::Debug(string format, ...args);
 * write debug log, only debug level print it.
 * @param format the format string
 * @param args the arguments list for format.
 */
func (self *Logger) Debug(f string, args...interface{}) {
	var fff = "[DEBUG] " + f + "\n";
	if (self.console) {
		fmt.Fprintf(os.Stdout, fff, args...);
	}
	if self.level >= DEBUG {
		self.logObj.Printf(fff, args...);
	}
}

/**
 * void Logger::Info(string format, ...args);
 * write info log, only INFO and DEBUG level print it.
 * @param format the format string
 * @param args the arguments list for format.
 */
func (self *Logger) Info(f string, args...interface{}) {
	var fff = "[INFO] " + f + "\n";
	if (self.console) {
		fmt.Fprintf(os.Stdout, fff, args...);
	}
	if self.level >= INFO {
		self.logObj.Printf(fff, args...);
	}
}

/**
 * void Logger::Warn(string format, ...args);
 * write warn log, only INFO and DEBUG, WARN level print it.
 * @param format the format string
 * @param args the arguments list for format.
 */
func (self *Logger) Warn(f string, args...interface{}) {
	var fff = "[WARN] " + f + "\n";
	if (self.console) {
		fmt.Fprintf(os.Stderr, fff, args...);
	}
	if self.level >= WARN {
		self.logObj.Printf(fff, args...);
	}
}

/**
 * void Logger::Err(string format, ...args);
 * write error log, only WARN, ERR, INFO and DEBUG level print it.
 * @param format the format string
 * @param args the arguments list for format.
 */
func (self *Logger) Err(f string, args...interface{}) {
	var fff = "[ERR] " + f + "\n";
	if (self.console) {
		fmt.Fprintf(os.Stderr, fff, args...);
	}
	if self.level >= ERR {
		self.logObj.Printf(fff, args...);
	}
}

/**
 * void Logger::Alert(string format, ...args);
 * write Alert log, only Fatal level NOT print it.
 * @param format the format string
 * @param args the arguments list for format.
 */
func (self *Logger) Alert(f string, args...interface{}) {
	var fff = "[ALERT] " + f + "\n";
	if (self.console) {
		fmt.Fprintf(os.Stderr, fff, args...);
	}
	if self.level >= ALERT {
		self.logObj.Printf(fff, args...);
	}
}

/**
 * void Logger::Fatal(string format, ...args);
 * write Fatal log, all log level print it.
 * NOTICE that this function will call os.Exit!
 * NOTICE that this function will call os.Exit!!
 * NOTICE that this function will call os.Exit!!!
 * @param format the format string
 * @param args the arguments list for format.
 */
func (self *Logger) Fatal(f string, args...interface{}) {
	var fff = "[FATAL] " + f + "\n";
	if (self.console) {
		fmt.Fprintf(os.Stderr, fff, args...);
	}
	if self.level >= FATAL {
		self.logObj.Fatalf(fff, args...);
	}
}

func (self *Logger) Write(buf []byte) (int, error) {
	self.lock.Lock();
	defer self.lock.Unlock();
	curTime := time.Now();

	// TODO: Notice that here we should make a copy of the buf, for detial info, please visit:
	// 		 https://go-zh.org/pkg/io/#Writer
	var sbuf string = fmt.Sprintf("%s %s", FormatLogTime(&curTime), *(*string)(unsafe.Pointer(&buf)));
	// rec := &logRecord{curTime, sbuf};

	// check rotate here
	if (curTime.Day() != self.flushTime.Day()) {
		// ROTATE log
		self.flushCache();
		self.logfile.Close();
		self.backupLog();
		self.open();
		self.flushTime = curTime;
	}
	if !GetCached() {
		// self.logfile.Write(buf);
		fmt.Fprintf(self.logfile,"%s", sbuf);
		self.flushTime = curTime;
		return len(buf), nil;
	}
	self.cache =  append(self.cache, sbuf);
	if (len(self.cache) >= BUF_SIZE) || (curTime.Unix() - self.flushTime.Unix() >= BUF_EXPIRES)  {
		self.flushCache();
		self.flushTime = curTime;
	}
	return len(buf), nil;
}

func FormatLogTime(rec *time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d.%03d", rec.Year(),
		rec.Month(), rec.Day(), rec.Hour(), rec.Minute(),
		rec.Second(), int(rec.Nanosecond() / 1000000));
}

func (rec *logRecord) String() string {
	// var s string = (*(*string)(unsafe.Pointer(&rec.data)));
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d.%03d %s", rec.logtime.Year(),
		rec.logtime.Month(), rec.logtime.Day(), rec.logtime.Hour(), rec.logtime.Minute(),
		rec.logtime.Second(), int(rec.logtime.Nanosecond() / 1000000), rec.data);
}

func (self *Logger) open() {
	var err error;
	var fname = path.Join(GetLogDir(), self.name + ".log");
	err = os.MkdirAll(GetLogDir(), 0755);
	if err != nil {
		fmt.Println(os.Stderr, "Err: %s\n", err);
		os.Exit(-1);
	}
	self.logfile, err = os.OpenFile(fname, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644);
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s\n", fname, err);
		os.Exit(-1);
	}
}

func (self *Logger) backupLog() {
	var oldName = path.Join(GetLogDir(), self.name + ".log");
	var newName = fmt.Sprintf("%s.%d-%02d-%02d.%d.log", self.name, self.flushTime.Year(),
		self.flushTime.Month(), self.flushTime.Day(), self.flushTime.Unix());
	newPath := path.Join(GetLogDir(), newName);
	os.Rename(oldName, newPath);
	if self.zipped {
		self.zipping = true;
		go zipLog(newName);
		self.zipping = false;
		select {
		case <- self.zipChan:
			self.zipChan <-1;
		default:
			self.zipChan <-1;
		}
	}
}

func newLogger(name string) *Logger {
	logger := new(Logger);
	logger.flushTime = time.Now();
	logger.level     = INFO;
	logger.name      = name;
	logger.console   = false;
	logger.zipped    = true;
	logger.zipping   = false;
	logger.zipChan   = make(chan int, 1);
	logger.logObj    = log.New(logger, "", 0);
	_, err := os.Stat(path.Join(GetLogDir(), logger.name+".log"));
	if err == nil {
		logger.backupLog();
	}
	logger.open();
	// logger.cache = make([]*logRecord, 0, BUF_SIZE);
	logger.cache = make([]string, 0, BUF_SIZE);
	/* runtime.SetFinalizer(logger, func(self *Logger) {
		fmt.Fprintf(os.Stdout, "log destroy...");
		self.flushCache();
		// self.logfile.Close();
	}) */
	return logger;
}

func (self *Logger) flushCache() {
	// fmt.Fprint(self.logfile, self.cache);
	// for _, rec := range self.cache {
	//	fmt.Fprint(self.logfile, rec);
	// }
	fmt.Fprint(self.logfile, strings.Join(self.cache, ""));
	self.cache = self.cache[0:0];
}

func zipLog(name string) {
	var err error;
	var oldDir string;
	// switch to log directory
	oldDir, err = os.Getwd();
	if err != nil {
		fmt.Fprintln(os.Stderr,"get current dir failed");
		return;
	}
	os.Chdir(GetLogDir());

	// create zip file
	var dstFd *os.File;
	dstFd, err = os.OpenFile(name + ".zip", os.O_CREATE, 0644);
	if err != nil {
		fmt.Fprintln(os.Stderr,"open zip file failed.");
		return;
	}

	// bind the zip file to the zip writer
	var zipWriter = zip.NewWriter(dstFd);

	// add a file to zip package, and get this files's writer
	var zipfileItemWriter, errx = zipWriter.Create(name);
	if errx != nil {
		fmt.Fprintln(os.Stderr, "add zip file failed");
		return;
	}

	// writer the file to the zip package's file
	var origFd *os.File;
	origFd, err = os.OpenFile(name, os.O_RDONLY, 444);
	if err != nil {
		fmt.Fprintln(os.Stderr,"open rotated file failed.");
		return;
	}
	var fbuf []byte = make([]byte, 40960, 40960);
	for {
		var rlen, e = origFd.Read(fbuf);
		if e != nil || rlen == 0 { break; }
		fbuf = fbuf[0:rlen];
		zipfileItemWriter.Write(fbuf);
	}

	origFd.Close();
	zipWriter.Flush();
	zipWriter.Close();
	dstFd.Close();
	err = os.Remove(name);
	if err != nil {
		fmt.Println(err);
	}
	os.Chdir(oldDir);
}