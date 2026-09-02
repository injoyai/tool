var EditorSetup = (function () {

    // ===== 默认 Go 脚本模板 =====
    var DEFAULT_SCRIPT = '' +
        'package main\n\n' +
        'import "fmt"\n\n' +
        'func main() {\n' +
        '    fmt.Println("hello timer")\n' +
        '}';

    // 错误处理脚本默认模板
    var ERROR_HANDLER_SCRIPT = '' +
        'package main\n\n' +
        'import "fmt"\n\n' +
        '// 任务执行失败时被调用\n' +
        '// taskID 稳定不变, taskName 可能被改名\n' +
        'func OnError(taskID int64, taskName, errMsg string) {\n' +
        '    fmt.Println("任务失败: "+taskName, errMsg)\n' +
        '}';

    // ===== 自动补全 =====
    // Go 关键字
    var GO_KEYWORDS = [
        'break', 'case', 'chan', 'const', 'continue', 'default', 'defer', 'else',
        'fallthrough', 'for', 'func', 'go', 'goto', 'if', 'import', 'interface',
        'map', 'package', 'range', 'return', 'select', 'struct', 'switch', 'type', 'var'
    ];

    // Go 内置类型与函数
    var GO_BUILTINS = [
        'append', 'cap', 'close', 'copy', 'delete', 'len', 'make', 'new',
        'panic', 'recover', 'error', 'complex', 'real', 'imag',
        'bool', 'byte', 'complex64', 'complex128', 'float32', 'float64',
        'int', 'int8', 'int16', 'int32', 'int64', 'rune', 'string',
        'uint', 'uint8', 'uint16', 'uint32', 'uint64', 'uintptr',
        'true', 'false', 'nil', 'iota', 'print', 'println'
    ];

    // Go 内置函数签名
    var GO_BUILTIN_SIGS = {
        'append': 'append(slice []Type, elems ...Type) []Type',
        'cap': 'cap(v Type) int',
        'close': 'close(c chan<- Type)',
        'copy': 'copy(dst, src []Type) int',
        'delete': 'delete(m map[Type]Type, key Type)',
        'len': 'len(v Type) int',
        'make': 'make(t Type, size ...Integer) Type',
        'new': 'new(Type) *Type',
        'panic': 'panic(v interface{})',
        'recover': 'recover() interface{}',
        'complex': 'complex(r, i float64) complex128',
        'real': 'real(c complex128) float64',
        'imag': 'imag(c complex128) float64',
        'print': 'print(args ...Type)',
        'println': 'println(args ...Type)'
    };

    // 包成员补全 (pkg. 后提示)
    var GO_HINTS = {
        'i': ['Start', 'Ping', 'Notice', 'Dial', 'DialTCP', 'Print', 'Println', 'ServerChan', 'QuarkCheckin', 'GetPrices', 'Set', 'Get', 'Del'],
        'fmt': ['Print', 'Println', 'Printf', 'Sprint', 'Sprintf', 'Sprintln', 'Errorf', 'Fprint', 'Fprintln', 'Fprintf', 'Scan', 'Scanln', 'Scanf', 'Sscan', 'Sscanln', 'Sscanf'],
        'time': ['Now', 'Second', 'Minute', 'Hour', 'Millisecond', 'Microsecond', 'Nanosecond', 'Sleep', 'Since', 'Until', 'Parse', 'ParseDuration', 'Duration', 'Time', 'Ticker', 'Timer', 'Date', 'Unix', 'UnixNano', 'Format', 'Local', 'UTC'],
        'os': ['Args', 'Stdin', 'Stdout', 'Stderr', 'Getenv', 'Setenv', 'Exit', 'Open', 'Create', 'ReadFile', 'WriteFile', 'MkdirAll', 'Remove', 'RemoveAll', 'Stat', 'IsExist', 'IsNotExist', 'Hostname', 'Getwd'],
        'net': ['Dial', 'DialTimeout', 'Listen', 'ResolveTCPAddr', 'ResolveIPAddr', 'SplitHostPort', 'JoinHostPort', 'ParseIP', 'IPv4'],
        'strings': ['Contains', 'ContainsAny', 'HasPrefix', 'HasSuffix', 'Index', 'IndexAny', 'LastIndex', 'Join', 'Replace', 'ReplaceAll', 'Split', 'SplitN', 'SplitAfter', 'Fields', 'Trim', 'TrimSpace', 'TrimLeft', 'TrimRight', 'ToUpper', 'ToLower', 'Title', 'Repeat', 'Count', 'Compare'],
        'strconv': ['Atoi', 'Itoa', 'ParseFloat', 'ParseInt', 'ParseUint', 'ParseBool', 'FormatFloat', 'FormatInt', 'FormatUint', 'AppendInt', 'AppendFloat'],
        'json': ['Marshal', 'Unmarshal', 'NewDecoder', 'NewEncoder', 'Indent', 'Compact', 'Valid', 'HTMLMarshal'],
        'regexp': ['MustCompile', 'Compile', 'Match', 'MatchString', 'MatchReader', 'QuoteMeta'],
        'errors': ['New', 'Is', 'As', 'Unwrap'],
        'sync': ['Mutex', 'RWMutex', 'WaitGroup', 'Once', 'Map', 'Pool', 'Cond'],
        'math': ['Abs', 'Ceil', 'Floor', 'Sqrt', 'Pow', 'Max', 'Min', 'Mod', 'Pi', 'Sin', 'Cos', 'Tan', 'Log', 'Exp', 'Rand', 'Inf', 'NaN', 'Trunc', 'Round'],
        'io': ['Reader', 'Writer', 'Copy', 'CopyN', 'ReadAll', 'ReadFull', 'WriteString', 'MultiReader', 'MultiWriter', 'Pipe', 'EOF'],
        'bufio': ['NewReader', 'NewWriter', 'NewScanner', 'ReadString', 'ReadBytes', 'ReadLine', 'WriteString', 'WriteBytes', 'Split', 'Scan', 'Text', 'Buffer'],
        'filepath': ['Join', 'Base', 'Dir', 'Ext', 'Clean', 'Abs', 'Rel', 'Walk', 'Glob', 'HasPrefix', 'FromSlash', 'ToSlash'],
        'sort': ['Sort', 'Ints', 'Strings', 'Float64s', 'IntsAreSorted', 'StringsAreSorted', 'Search', 'SearchInts', 'SearchStrings', 'Slice', 'SliceStable', 'Reverse'],
        'context': ['Background', 'TODO', 'WithCancel', 'WithTimeout', 'WithDeadline', 'WithValue', 'Cancel'],
        'bytes': ['NewBuffer', 'NewBufferString', 'NewReader', 'Compare', 'Contains', 'HasPrefix', 'HasSuffix', 'Index', 'Join', 'Replace', 'Split', 'Trim', 'Reader', 'Buffer'],
        'ioutil': ['ReadFile', 'WriteFile', 'ReadAll', 'ReadDir', 'TempFile', 'TempDir', 'NopCloser', 'Discard'],
        'unicode': ['IsLetter', 'IsDigit', 'IsSpace', 'IsPrint', 'IsUpper', 'IsLower', 'IsPunct', 'IsControl', 'IsMark', 'IsNumber', 'IsSymbol', 'ToUpper', 'ToLower', 'ToTitle'],
        'reflect': ['TypeOf', 'ValueOf', 'New', 'Zero', 'DeepEqual', 'Copy', 'MakeSlice', 'MakeMap', 'MakeChan'],
        // injoyai 库
        'conv': ['Array', 'BIN', 'BINBool', 'BINStr', 'Bool', 'Byte', 'Bytes', 'BytesZ', 'Copy', 'DMap', 'Duration', 'Float32', 'Float64', 'GMap', 'HEX', 'HEXStr', 'IMap', 'Int', 'Int16', 'Int32', 'Int64', 'Int64s', 'Int8', 'Interfaces', 'Ints', 'IsArray', 'IsBool', 'IsDefault', 'IsFloat', 'IsInt', 'IsNil', 'IsNumber', 'IsPointer', 'IsString', 'IsTime', 'IsZero', 'New', 'NewExtend', 'NewMap', 'Nil', 'OCT', 'OCTStr', 'Rune', 'Runes', 'SMap', 'String', 'Strings', 'Uint', 'Uint16', 'Uint32', 'Uint64', 'Uint8', 'Unmarshal'],
        'logs': ['AddWriter', 'Debug', 'Debugf', 'Err', 'Errf', 'Error', 'Errorf', 'Fatal', 'Fatalf', 'Info', 'Infof', 'LevelAll', 'LevelDebug', 'LevelError', 'LevelInfo', 'LevelNone', 'LevelRead', 'LevelTrace', 'LevelWarn', 'LevelWrite', 'New', 'NewFile', 'Panic', 'PanicErr', 'Panicf', 'PrintErr', 'Read', 'Readf', 'SetFormatter', 'SetLevel', 'SetShowColor', 'SetWriter', 'Spend', 'Stdout', 'Trace', 'Tracef', 'Warn', 'Warnf', 'Write', 'Writef'],
        'str': ['Bool', 'Bytes', 'Contains', 'Count', 'CropFirst', 'CropLast', 'FindCommon', 'Float32', 'Float64', 'GB18030ToUtf8', 'GbkToUtf8', 'GetLine', 'GetSplit', 'HasPrefix', 'HasRepeat', 'HasSuffix', 'HZGB2312ToUtf8', 'Index', 'Int', 'Int16', 'Int32', 'Int64', 'Int8', 'Interface', 'IsBegin', 'IsEnd', 'Join', 'MustSplitN', 'NewReader', 'Pointer', 'Rand', 'ReplaceAll', 'Reverse', 'Select', 'Split', 'Title', 'ToLower', 'ToUpper', 'TrimPrefix', 'TrimSpace', 'TrimSuffix', 'Uint16', 'Uint32', 'Uint64', 'Uint8'],
        'crypt': ['DealLength', 'Hmac', 'New', 'Padding', 'Entity'],
        'coding': ['DecodeASCII', 'DecodeBase64', 'DecodeHEX', 'EncodeASCII', 'EncodeBase64', 'EncodeHEX', 'JsonMarshal', 'JsonUnmarshal', 'MsgpackMarshal', 'MsgpackUnmarshal', 'ProtoMarshal', 'ProtoUnmarshal', 'TomlMarshal', 'TomlUnmarshal', 'XmlMarshal', 'XmlUnmarshal', 'YamlMarshal', 'YamlUnmarshal'],
        'maps': ['NewBit', 'NewSafe', 'Bit'],
        'safe': ['Index', 'NewCloser', 'NewCloserErr', 'NewGoroute', 'NewGorouteWithContext', 'NewInt32', 'NewInt64', 'NewOneRun', 'NewRerun', 'NewRunner', 'NewRunner2', 'NewRunnerWithContext', 'NewUint32', 'NewUint64', 'NewUse', 'Recover', 'RecoverFunc', 'Retry', 'Try', 'Bool', 'Closer', 'Dialer', 'Goroute', 'Int32', 'Int64', 'Once', 'OneRun', 'Rerun', 'Runner', 'Runner2', 'Uint32', 'Uint64', 'Use', 'Value'],
        'types': ['NewMemory', 'Sort', 'Bs', 'Bytes', 'Closer', 'Debugger', 'Doner', 'Err', 'F32', 'F64', 'Float', 'Float32', 'Float64', 'Message', 'Price', 'Runner', 'Signaler', 'Sorter'],
        'chans': ['NewCoroutine', 'NewCounter', 'NewIO', 'NewLast', 'NewLimit', 'NewLimitGo', 'NewOrder', 'NewQueueFunc', 'NewRerun', 'NewSignal', 'NewWaitLimit', 'TraverseInterval', 'Coroutine', 'Counter', 'IO', 'Last', 'Limit', 'LimitGo', 'Order', 'OrderNode', 'QueueFunc', 'QueueFuncHandler', 'Rerun', 'Signal', 'WaitLimit'],
        // ios/v2 包
        'ios': ['Bridge', 'DefaultBufferSize', 'Discard', 'ErrClosed', 'ErrReadTimeout', 'ErrRemoteClose', 'ErrWriteTimeout', 'MultiCloser', 'NewAllReader', 'NewBuffer', 'NewFRead', 'NewFRead4KB', 'NewFReadB', 'NewFReadKB', 'NewFReadLeast', 'NewMoreWrite', 'NewPiper', 'NewPlanWrite', 'Null', 'Pipe', 'ReadByte', 'ReadPrefix', 'Buffer', 'Bytes', 'ChanWriter', 'Closer', 'Listener', 'Piper', 'Reader', 'Writer', 'ReadCloser', 'ReadWriteCloser'],
        'client': ['DefaultReaderPool', 'Dial', 'DialContext', 'New', 'NewDealMessageWithChan', 'NewDealMessageWithWriter', 'NewDisconnectAfter', 'NewPool', 'NewReconnectInterval', 'NewReconnectRetreat', 'NewWriteRetry', 'NewWriteSafe', 'Redial', 'RedialContext', 'Run', 'RunContext', 'WithConnect', 'WithDealMessage', 'WithDebug', 'WithDisconnect', 'WithFrame', 'WithHEX', 'WithLevel', 'WithReadFrom', 'WithRedial', 'WithUTF8', 'WithWriteRetry', 'WithWriteSafe', 'WithWriteWith', 'Client', 'Frame', 'Info', 'Option', 'Pool'],
        'dial': ['Memory', 'Run', 'RunMemory', 'RunTCP', 'RunUDP', 'RunUnix', 'RunWebsocket', 'TCP', 'UDP', 'Unix', 'Websocket', 'With'],
        'frame': ['Default', 'Prefix', 'ReadFrom', 'WriteWith', 'Frame'],
        'redial': ['Memory', 'Run', 'RunMemory', 'RunTCP', 'RunUDP', 'RunUnix', 'RunWebsocket', 'TCP', 'UDP', 'Unix', 'Websocket', 'With'],
        'server': ['New', 'Run', 'RunContext', 'WithClientConnected', 'WithClientOptions', 'WithLoggerDisable', 'WithLoggerEnable', 'WithLoggerLevel', 'Event', 'Option', 'Server'],
        'listen': ['Memory', 'Run', 'RunMemory', 'RunUnix', 'Unix'],
        'split': ['CRC16Modbus', 'Checker', 'Length', 'Prefix', 'Prefixes', 'Regular', 'Split', 'Suffix', 'SumLast'],
        'common': ['DealErr', 'LevelAll', 'LevelDebug', 'LevelError', 'LevelInfo', 'LevelNone', 'NewLogger', 'Logger'],
        'memory': ['Dial', 'DialTimeout', 'NewDial', 'NewListen', 'Client', 'Server'],
        'mqtt': ['Dial', 'DialClient', 'NewDial', 'NewListen', 'NewNetListen', 'WithBase', 'BaseConfig', 'Client', 'Config', 'Conn', 'Connect', 'Message', 'Publish', 'Server', 'Subscribe'],
        'serial': ['Dial', 'NewDial', 'Open', 'Client', 'Config', 'RS485Config'],
        'sse': ['Dial', 'NewDial', 'NewHandlerListen', 'NewListen', 'Args', 'Client', 'Server'],
        'ssh': ['Dial', 'NewDial', 'Client', 'Config'],
        'tcp': ['NewDial', 'Server'],
        'udp': ['NewDial', 'Server'],
        'unix': ['NewDial', 'NewListen', 'Server'],
        'websocket': ['Dial', 'NewDial', 'NewNetListen', 'Client', 'Conn', 'Dialer', 'Server'],
        // 其他 injoyai 包
        'bar': ['Animations', 'Copy', 'DefaultPadding', 'DefaultStyle', 'Download', 'DownloadHLS', 'New', 'NewCoroutine', 'NewPlan', 'Stat', 'WithAnimation', 'WithAutoFlush', 'WithCurrent', 'WithDate', 'WithDateTime', 'WithFinal', 'WithFlush', 'WithFormat', 'WithIntervalFlush', 'WithOption', 'WithPlan', 'WithPrefix', 'WithRate', 'WithRateSize', 'WithRemain', 'WithSpeed', 'WithSuffix', 'WithText', 'WithTime', 'WithTotal', 'WithUsed', 'WithWriter', 'Bar', 'Coroutine', 'Format', 'Option', 'Plan', 'Reader'],
        'aes': ['DecryptCBC', 'DecryptCBCBase64', 'DecryptCBCBytes', 'DecryptCBCHEX', 'DecryptCBCString', 'DecryptECB', 'DecryptECBBase64', 'DecryptECBBytes', 'DecryptECBHEX', 'DecryptECBString', 'EncryptCBC', 'EncryptCBCBase64', 'EncryptCBCBytes', 'EncryptCBCHEX', 'EncryptCBCString', 'EncryptECB', 'EncryptECBBase64', 'EncryptECBBytes', 'EncryptECBHEX', 'EncryptECBString', 'PKCS7Padding', 'PKCS7UnPadding'],
        'crc': ['CRC16_MODBUS', 'CRC16_USB', 'CRC16_XMODEM', 'CRC8', 'Checksum16', 'Checksum8', 'Complete16', 'Complete8', 'Encrypt16', 'Encrypt16Base64', 'Encrypt16HEX', 'Encrypt16String', 'Encrypt8', 'Init16', 'Init8', 'MakeTable16', 'MakeTable8', 'Table16', 'Table8'],
        'des': ['DecryptECB', 'DecryptECBBase64', 'DecryptECBBytes', 'DecryptECBHEX', 'DecryptECBString', 'EncryptECB', 'EncryptECBBase64', 'EncryptECBBytes', 'EncryptECBHEX', 'EncryptECBString'],
        'gzip': ['DecodeGzip', 'EncodeGzip'],
        'md5': ['Encrypt', 'EncryptBase64', 'EncryptBytes', 'EncryptHEX', 'EncryptString', 'Hmac', 'HmacBase64', 'HmacBytes', 'HmacHEX', 'HmacString'],
        'sha': ['Encrypt1', 'Encrypt1Base64', 'Encrypt1Bytes', 'Encrypt1HEX', 'Encrypt1String', 'Encrypt256', 'Encrypt256Base64', 'Encrypt256Bytes', 'Encrypt256HEX', 'Encrypt256String', 'Encrypt512', 'Encrypt512Base64', 'Encrypt512Bytes', 'Encrypt512HEX', 'Encrypt512String', 'Hmac1', 'Hmac256', 'Hmac512'],
        'tls': ['Config'],
        'timeout': ['New'],
        'wait': ['Async', 'Default', 'Done', 'IsWait', 'New', 'SetReuse', 'SetTimeout', 'Sync', 'Wait'],
        'cfg': ['Append', 'Default', 'GetBool', 'GetDuration', 'GetFloat64', 'GetInt', 'GetInt64', 'GetInts', 'GetMap', 'GetString', 'GetStrings', 'Init', 'New', 'WithEnv', 'WithFile', 'WithJson', 'WithYaml', 'Entity', 'Env'],
        'codec': ['Default', 'Get', 'Ini', 'Json', 'Toml', 'Yaml'],
        'ini': ['Ini'],
        'toml': ['Toml'],
        'xml': ['Xml'],
        'yaml': ['Yaml'],
        'fbr': ['BindCode', 'BindHtml', 'Default', 'New', 'NewCtx', 'WithCORS', 'WithGET', 'WithPOST', 'WithPort', 'WithRecover', 'WithStatic', 'Bind', 'Ctx', 'Grouper', 'Handler', 'Option', 'Server', 'Websocket'],
        'gins': ['BindCode', 'BindHtml', 'Default', 'New', 'WithCORS', 'WithPort', 'WithRecover', 'WithSwagger', 'Ctx', 'Grouper', 'Handler', 'Option', 'Server'],
        'swagger': ['Default', 'DefaultUI', 'Swagger'],
        'frame': ['DefaultPort', 'NewLogger', 'Logger']
    };

    // 函数签名 (补全列表显示签名,选中只插入函数名)
    var GO_SIGS = {
        'i': {
            'Start': 'Start(cmd string) error',
            'Ping': 'Ping(host string, timeout time.Duration) (string, error)',
            'Notice': 'Notice(msg string) error',
            'Dial': 'Dial(network, address string, timeout time.Duration) (string, error)',
            'DialTCP': 'DialTCP(address string, timeout time.Duration) (string, error)',
            'Print': 'Print(args ...interface{})',
            'Println': 'Println(args ...interface{})',
            'ServerChan': 'ServerChan(title, msg string) error',
            'QuarkCheckin': 'QuarkCheckin(vcode, sign, kps string) (string, error)',
            'GetPrices': 'GetPrices(code ...string) (map[string]float64, error)',
            'Set': 'Set(key, value any, expiration ...time.Duration)',
            'Get': 'Get(key any) (any, bool)',
            'Del': 'Del(key any)'
        },
        'fmt': {
            'Print': 'Print(a ...any) (n int, err error)',
            'Println': 'Println(a ...any) (n int, err error)',
            'Printf': 'Printf(format string, a ...any) (n int, err error)',
            'Sprint': 'Sprint(a ...any) string',
            'Sprintf': 'Sprintf(format string, a ...any) string',
            'Sprintln': 'Sprintln(a ...any) string',
            'Errorf': 'Errorf(format string, a ...any) error',
            'Fprint': 'Fprint(w io.Writer, a ...any) (n int, err error)',
            'Fprintln': 'Fprintln(w io.Writer, a ...any) (n int, err error)',
            'Fprintf': 'Fprintf(w io.Writer, format string, a ...any) (n int, err error)'
        },
        'time': {
            'Now': 'Now() Time',
            'Sleep': 'Sleep(d Duration)',
            'Since': 'Since(t Time) Duration',
            'Until': 'Until(t Time) Duration',
            'Parse': 'Parse(layout, value string) (Time, error)',
            'ParseDuration': 'ParseDuration(s string) (Duration, error)',
            'Date': 'Date(year int, month Month, day, hour, min, sec, nsec int, loc *Location) Time',
            'Unix': 'Unix(sec int64, nsec int64) Time',
            'Format': '(t Time) Format(layout string) string'
        },
        'os': {
            'Getenv': 'Getenv(key string) string',
            'Setenv': 'Setenv(key, value string) error',
            'Exit': 'Exit(code int)',
            'Open': 'Open(name string) (*File, error)',
            'Create': 'Create(name string) (*File, error)',
            'ReadFile': 'ReadFile(name string) ([]byte, error)',
            'WriteFile': 'WriteFile(name string, data []byte, perm FileMode) error',
            'MkdirAll': 'MkdirAll(path string, perm FileMode) error',
            'Remove': 'Remove(name string) error',
            'RemoveAll': 'RemoveAll(path string) error',
            'Stat': 'Stat(name string) (FileInfo, error)',
            'Getwd': 'Getwd() (string, error)'
        },
        'strings': {
            'Contains': 'Contains(s, substr string) bool',
            'HasPrefix': 'HasPrefix(s, prefix string) bool',
            'HasSuffix': 'HasSuffix(s, suffix string) bool',
            'Index': 'Index(s, substr string) int',
            'Join': 'Join(e []string, sep string) string',
            'Replace': 'Replace(s, old, new string, n int) string',
            'ReplaceAll': 'ReplaceAll(s, old, new string) string',
            'Split': 'Split(s, sep string) []string',
            'SplitN': 'SplitN(s, sep string, n int) []string',
            'ToUpper': 'ToUpper(s string) string',
            'ToLower': 'ToLower(s string) string',
            'Trim': 'Trim(s, cutset string) string',
            'TrimSpace': 'TrimSpace(s string) string',
            'Count': 'Count(s, substr string) int',
            'Repeat': 'Repeat(s string, n int) string'
        },
        'strconv': {
            'Atoi': 'Atoi(s string) (int, error)',
            'Itoa': 'Itoa(i int) string',
            'ParseFloat': 'ParseFloat(s string, bitSize int) (float64, error)',
            'ParseInt': 'ParseInt(s string, base, bitSize int) (int64, error)',
            'ParseBool': 'ParseBool(str string) (bool, error)',
            'FormatFloat': 'FormatFloat(f float64, fmt byte, prec, bitSize int) string',
            'FormatInt': 'FormatInt(i int64, base int) string'
        },
        'json': {
            'Marshal': 'Marshal(v any) ([]byte, error)',
            'Unmarshal': 'Unmarshal(data []byte, v any) error',
            'NewDecoder': 'NewDecoder(r io.Reader) *Decoder',
            'NewEncoder': 'NewEncoder(w io.Writer) *Encoder'
        },
        'regexp': {
            'MustCompile': 'MustCompile(str string) *Regexp',
            'Compile': 'Compile(str string) (*Regexp, error)',
            'MatchString': 'MatchString(pattern, s string) (bool, error)'
        },
        'conv': {
            'Int': 'Int(v interface{}) int',
            'Int64': 'Int64(v interface{}) int64',
            'String': 'String(v interface{}) string',
            'Bool': 'Bool(v interface{}) bool',
            'Float64': 'Float64(v interface{}) float64',
            'Bytes': 'Bytes(v interface{}) []byte',
            'Duration': 'Duration(v interface{}) time.Duration',
            'Interfaces': 'Interfaces(v interface{}) []interface{}',
            'Ints': 'Ints(v interface{}) []int',
            'Strings': 'Strings(v interface{}) []string',
            'Unmarshal': 'Unmarshal(b []byte, ptr interface{}) error'
        },
        'logs': {
            'Debug': 'Debug(v ...interface{})',
            'Debugf': 'Debugf(format string, v ...interface{})',
            'Info': 'Info(v ...interface{})',
            'Infof': 'Infof(format string, v ...interface{})',
            'Warn': 'Warn(v ...interface{})',
            'Warnf': 'Warnf(format string, v ...interface{})',
            'Err': 'Err(v ...interface{})',
            'Errf': 'Errf(format string, v ...interface{})',
            'Printf': 'Printf(format string, v ...interface{})',
            'Println': 'Println(v ...interface{})'
        },
        'str': {
            'Contains': 'Contains(s, substr string) bool',
            'HasPrefix': 'HasPrefix(s, prefix string) bool',
            'HasSuffix': 'HasSuffix(s, suffix string) bool',
            'Join': 'Join(a []string, sep string) string',
            'Split': 'Split(s, sep string) []string',
            'ReplaceAll': 'ReplaceAll(s, old, new string) string',
            'ToUpper': 'ToUpper(s string) string',
            'ToLower': 'ToLower(s string) string',
            'TrimSpace': 'TrimSpace(s string) string',
            'Int': 'Int(v interface{}) int',
            'Int64': 'Int64(v interface{}) int64',
            'Float64': 'Float64(v interface{}) float64',
            'Bool': 'Bool(v interface{}) bool'
        },
        'net': {
            'Dial': 'Dial(network, address string) (Conn, error)',
            'DialTimeout': 'DialTimeout(network, address string, timeout time.Duration) (Conn, error)',
            'Listen': 'Listen(network, address string) (Listener, error)',
            'ParseIP': 'ParseIP(s string) IP'
        },
        // ios/v2
        'ios': {
            'Bridge': 'Bridge(r1, r2 io.ReadWriteCloser) error',
            'MultiCloser': 'MultiCloser(closer ...io.Closer) io.Closer',
            'Pipe': 'Pipe(cap int) (BReadWriteCloser, BReadWriteCloser)',
            'NewPiper': 'NewPiper(cap int) *Piper',
            'NewBuffer': 'NewBuffer(r io.Reader, buf []byte) *Buffer',
            'NewFRead': 'NewFRead(buf []byte) FReadFunc',
            'NewFReadLeast': 'NewFReadLeast(least int) FReadFunc',
            'NewFReadB': 'NewFReadB(n int) FReadFunc',
            'NewFReadKB': 'NewFReadKB(n int) FReadFunc',
            'NewFRead4KB': 'NewFRead4KB() FReadFunc',
            'ReadByte': 'ReadByte(r io.Reader) (byte, error)',
            'ReadPrefix': 'ReadPrefix(r io.Reader, prefix []byte) ([]byte, error)',
            'NewAllReader': 'NewAllReader(r Reader, f func(r io.Reader) ([]byte, error)) *AllRead'
        },
        'client': {
            'Run': 'Run(dial ios.DialFunc, op ...Option) error',
            'RunContext': 'RunContext(ctx context.Context, dial ios.DialFunc, op ...Option) error',
            'Redial': 'Redial(dial ios.DialFunc, op ...Option) *Client',
            'RedialContext': 'RedialContext(ctx context.Context, dial ios.DialFunc, op ...Option) *Client',
            'Dial': 'Dial(f ios.DialFunc, op ...Option) (*Client, error)',
            'DialContext': 'DialContext(ctx context.Context, dial ios.DialFunc, op ...Option) (*Client, error)',
            'New': 'New(dial ios.DialFunc, op ...Option) *Client',
            'NewPool': 'NewPool(max int, new func() *bufio.Reader) *Pool',
            'NewDealMessageWithChan': 'NewDealMessageWithChan(ch chan Acker) func(*Client, Acker)',
            'NewDealMessageWithWriter': 'NewDealMessageWithWriter(w io.Writer) func(*Client, Acker)',
            'NewDisconnectAfter': 'NewDisconnectAfter(t time.Duration) func(*Client, error) error'
        },
        'dial': {
            'TCP': 'TCP(addr string, op ...Option) (*Client, error)',
            'RunTCP': 'RunTCP(addr string, op ...Option) error',
            'UDP': 'UDP(addr string, op ...Option) (*Client, error)',
            'RunUDP': 'RunUDP(addr string, op ...Option) error',
            'Unix': 'Unix(addr string, op ...Option) (*Client, error)',
            'RunUnix': 'RunUnix(addr string, op ...Option) error',
            'Websocket': 'Websocket(addr string, op ...Option) (*Client, error)',
            'RunWebsocket': 'RunWebsocket(addr string, op ...Option) error',
            'Memory': 'Memory(key string, op ...Option) (*Client, error)',
            'RunMemory': 'RunMemory(key string, op ...Option) error'
        },
        'redial': {
            'TCP': 'TCP(addr string, op ...Option) *Client',
            'UDP': 'UDP(addr string, op ...Option) *Client',
            'Unix': 'Unix(addr string, op ...Option) *Client',
            'Websocket': 'Websocket(addr string, op ...Option) *Client',
            'Memory': 'Memory(key string, op ...Option) *Client'
        },
        'server': {
            'New': 'New(listen ios.ListenFunc, op ...Option) (*Server, error)',
            'Run': 'Run(listen ios.ListenFunc, op ...Option) error',
            'RunContext': 'RunContext(ctx context.Context, listen ios.ListenFunc, op ...Option) error'
        },
        'listen': {
            'TCP': 'TCP(addr string, op ...Option) (*Server, error)',
            'RunTCP': 'RunTCP(addr string, op ...Option) error',
            'UDP': 'UDP(addr string, op ...Option) (*Server, error)',
            'RunUDP': 'RunUDP(addr string, op ...Option) error',
            'Unix': 'Unix(filename string, op ...Option) (*Server, error)',
            'RunUnix': 'RunUnix(filename string, op ...Option) error',
            'Memory': 'Memory(key string, op ...Option) (*Server, error)',
            'RunMemory': 'RunMemory(key string, op ...Option) error',
            'Websocket': 'Websocket(addr string, op ...Option) (*Server, error)',
            'RunWebsocket': 'RunWebsocket(addr string, op ...Option) error'
        },
        'websocket': {
            'Dial': 'Dial(url string) (*Client, error)',
            'NewDial': 'NewDial(url string) ios.DialFunc',
            'NewListen': 'NewListen(addr string) func() (ios.Listener, error)',
            'NewNetListen': 'NewNetListen(l net.Listener) func() (ios.Listener, error)'
        },
        'tcp': {
            'NewDial': 'NewDial(addr string) ios.DialFunc',
            'NewListen': 'NewListen(addr string) func() (ios.Listener, error)'
        },
        'udp': {
            'NewDial': 'NewDial(addr string) ios.DialFunc',
            'NewListen': 'NewListen(addr string) func() (ios.Listener, error)'
        },
        'unix': {
            'NewDial': 'NewDial(addr string) ios.DialFunc',
            'NewListen': 'NewListen(filename string) ios.ListenFunc'
        },
        'memory': {
            'Dial': 'Dial(key string) (*Client, error)',
            'DialTimeout': 'DialTimeout(key string, timeout time.Duration) (*Client, error)',
            'NewDial': 'NewDial(key string) ios.DialFunc',
            'NewListen': 'NewListen(key string) func() (ios.Listener, error)'
        },
        'mqtt': {
            'Dial': 'Dial(c Connect, sub Subscribe, pub Publish) (*Client, error)',
            'DialClient': 'DialClient(cfg *Config, sub Subscribe, pub Publish) (*Client, error)',
            'NewDial': 'NewDial(cfg *Config, sub Subscribe, pub Publish) ios.DialFunc',
            'NewListen': 'NewListen(port int) ios.ListenFunc',
            'NewNetListen': 'NewNetListen(l net.Listener) ios.ListenFunc',
            'WithBase': 'WithBase(cfg *BaseConfig) *mqtt.ClientOptions'
        },
        'serial': {
            'Dial': 'Dial(cfg *Config) (Client, error)',
            'Open': 'Open(cfg *Config) (Client, error)',
            'NewDial': 'NewDial(cfg *Config) ios.DialFunc'
        },
        'ssh': {
            'Dial': 'Dial(cfg *Config) (*Client, error)',
            'NewDial': 'NewDial(cfg *Config) ios.DialFunc'
        },
        'sse': {
            'Dial': 'Dial(url string, body io.Reader) (*Client, error)',
            'NewDial': 'NewDial(url string, body io.Reader) ios.DialFunc',
            'NewListen': 'NewListen(port int) func() (ios.Listener, error)',
            'NewHandlerListen': 'NewHandlerListen(f func(h http.Handler)) func() (ios.Listener, error)'
        },
        'common': {
            'DealErr': 'DealErr(err error) error',
            'NewLogger': 'NewLogger() *logger'
        },
        'split': {
            'CRC16Modbus': 'CRC16Modbus struct{}',
            'Length': 'Length struct{ LittleEndian bool; Start, End uint; Fixed int }',
            'Prefix': 'Prefix []byte',
            'Suffix': 'Suffix []byte',
            'Regular': 'Regular struct{ *regexp.Regexp }'
        },
        'frame': {
            'ReadFrom': 'ReadFrom(r io.Reader) ([]byte, error)',
            'WriteWith': 'WriteWith(bs []byte) ([]byte, error)'
        },
        // bar
        'bar': {
            'New': 'New(op ...Option) *Bar',
            'WithTotal': 'WithTotal(total int64) Option',
            'WithCurrent': 'WithCurrent(current int64) Option',
            'WithWriter': 'WithWriter(writer io.Writer) Option',
            'WithPrefix': 'WithPrefix(prefix string) Option',
            'WithSuffix': 'WithSuffix(suffix string) Option',
            'WithFormat': 'WithFormat(fs ...Format) Option',
            'WithAutoFlush': 'WithAutoFlush() Option',
            'WithIntervalFlush': 'WithIntervalFlush(interval time.Duration) Option',
            'WithFinal': 'WithFinal(f Option) Option',
            'WithFinalLn': 'WithFinalLn() Option'
        },
        // crypt 子包
        'aes': {
            'EncryptCBC': 'EncryptCBC(bs, key []byte, ivs ...[]byte) (types.Bytes, error)',
            'EncryptCBCString': 'EncryptCBCString(bs, key []byte, iv ...[]byte) (string, error)',
            'EncryptCBCHEX': 'EncryptCBCHEX(bs, key []byte, iv ...[]byte) (string, error)',
            'EncryptCBCBase64': 'EncryptCBCBase64(bs, key []byte, iv ...[]byte) (string, error)',
            'DecryptCBC': 'DecryptCBC(bs, key []byte, ivs ...[]byte) (types.Bytes, error)',
            'DecryptCBCString': 'DecryptCBCString(bs, key []byte, ivs ...[]byte) (string, error)',
            'DecryptCBCHEX': 'DecryptCBCHEX(bs, key []byte, ivs ...[]byte) (string, error)',
            'DecryptCBCBase64': 'DecryptCBCBase64(bs, key []byte, ivs ...[]byte) (string, error)',
            'EncryptECB': 'EncryptECB(bs, key []byte) (types.Bytes, error)',
            'EncryptECBString': 'EncryptECBString(bs, key []byte) (string, error)',
            'DecryptECB': 'DecryptECB(bs, key []byte) (types.Bytes, error)',
            'DecryptECBString': 'DecryptECBString(bs, key []byte) (string, error)',
            'PKCS7Padding': 'PKCS7Padding(ciphertext []byte, blockSize int) []byte',
            'PKCS7UnPadding': 'PKCS7UnPadding(origData []byte) []byte'
        },
        'crc': {
            'Encrypt16': 'Encrypt16(bs []byte, params ...Param16) []byte',
            'Encrypt16String': 'Encrypt16String(bs []byte, params ...Param16) string',
            'Encrypt16HEX': 'Encrypt16HEX(bs []byte, params ...Param16) string',
            'Encrypt16Base64': 'Encrypt16Base64(bs []byte, params ...Param16) string',
            'Encrypt8': 'Encrypt8(bs []byte, params ...Param8) byte',
            'Checksum16': 'Checksum16(data []byte, table *Table16) uint16',
            'Checksum8': 'Checksum8(data []byte, table *Table8) uint8',
            'MakeTable16': 'MakeTable16(params Param16) *Table16',
            'MakeTable8': 'MakeTable8(params Param8) *Table8'
        },
        'des': {
            'EncryptECB': 'EncryptECB(str, key string) string',
            'EncryptECBBytes': 'EncryptECBBytes(str, key string) []byte',
            'EncryptECBHEX': 'EncryptECBHEX(str, key string) string',
            'EncryptECBBase64': 'EncryptECBBase64(str, key string) string',
            'DecryptECB': 'DecryptECB(str, key string) string',
            'DecryptECBBytes': 'DecryptECBBytes(str, key string) []byte',
            'DecryptECBHEX': 'DecryptECBHEX(str, key string) string',
            'DecryptECBBase64': 'DecryptECBBase64(str, key string) string'
        },
        'gzip': {
            'EncodeGzip': 'EncodeGzip(input []byte) ([]byte, error)',
            'DecodeGzip': 'DecodeGzip(input []byte) ([]byte, error)'
        },
        'md5': {
            'Encrypt': 'Encrypt(s string) string',
            'EncryptBytes': 'EncryptBytes(s string) []byte',
            'EncryptHEX': 'EncryptHEX(s string) string',
            'EncryptBase64': 'EncryptBase64(s string) string',
            'Hmac': 'Hmac(s string, key string) string',
            'HmacBytes': 'HmacBytes(s string, key string) []byte',
            'HmacHEX': 'HmacHEX(s string, key string) string',
            'HmacBase64': 'HmacBase64(s string, key string) string'
        },
        'sha': {
            'Encrypt1': 'Encrypt1(bs []byte) types.Bytes',
            'Encrypt1String': 'Encrypt1String(bs []byte) string',
            'Encrypt1HEX': 'Encrypt1HEX(bs []byte) string',
            'Encrypt1Base64': 'Encrypt1Base64(bs []byte) string',
            'Encrypt256': 'Encrypt256(bs []byte) types.Bytes',
            'Encrypt256String': 'Encrypt256String(bs []byte) string',
            'Encrypt256HEX': 'Encrypt256HEX(bs []byte) string',
            'Encrypt256Base64': 'Encrypt256Base64(bs []byte) string',
            'Encrypt512': 'Encrypt512(bs []byte) types.Bytes',
            'Encrypt512String': 'Encrypt512String(bs []byte) string',
            'Encrypt512HEX': 'Encrypt512HEX(bs []byte) string',
            'Encrypt512Base64': 'Encrypt512Base64(bs []byte) string',
            'Hmac1': 'Hmac1(data, secret []byte) types.Bytes',
            'Hmac256': 'Hmac256(data, secret []byte) types.Bytes',
            'Hmac512': 'Hmac512(data, secret []byte) types.Bytes'
        },
        'tls': {
            'Config': 'Config{ CAFile, CCFile, CKFile string }'
        },
        // maps 子包
        'timeout': {
            'New': 'New() *Timeout'
        },
        'wait': {
            'Wait': 'Wait(key string, timeout ...time.Duration) (any, error)',
            'Sync': 'Sync(key string, timeout ...time.Duration) (any, error)',
            'Async': 'Async(key string, f Handler[any], num int, timeout ...time.Duration)',
            'Done': 'Done(key string, v any, err ...error) bool',
            'IsWait': 'IsWait(key string) bool',
            'New': 'New(timeout time.Duration) *Entity',
            'SetTimeout': 'SetTimeout(t time.Duration) *Entity',
            'SetReuse': 'SetReuse(b ...bool) *Entity'
        },
        // conv 子包
        'cfg': {
            'GetString': 'GetString(key string, def ...string) string',
            'GetInt': 'GetInt(key string, def ...int) int',
            'GetInt64': 'GetInt64(key string, def ...int64) int64',
            'GetBool': 'GetBool(key string, def ...bool) bool',
            'GetFloat64': 'GetFloat64(key string, def ...float64) float64',
            'GetDuration': 'GetDuration(key string, def ...time.Duration) time.Duration',
            'GetString': 'GetString(key string, def ...string) string',
            'GetStrings': 'GetStrings(key string, def ...[]string) []string',
            'GetInts': 'GetInts(key string, def ...[]int) []int',
            'GetMap': 'GetMap(key string, def ...map[string]any) map[string]any',
            'Init': 'Init(i ...conv.IGetVar)',
            'New': 'New(i ...conv.IGetVar) *Entity',
            'WithEnv': 'WithEnv() conv.IGetVar',
            'WithFile': 'WithFile(filename string, codecs ...codec.Interface) conv.IGetVar',
            'WithYaml': 'WithYaml(filename string) conv.IGetVar',
            'WithJson': 'WithJson(filename string) conv.IGetVar'
        },
        'codec': {
            'Get': 'Get(s string) Interface'
        },
        // frame 子包
        'fbr': {
            'New': 'New(use ...Middle) *Server',
            'Default': 'Default(use ...Middle) *Server',
            'NewCtx': 'NewCtx(c fiber.Ctx, r Respondent) Ctx',
            'WithPort': 'WithPort(port int) Option',
            'WithCORS': 'WithCORS() Middle',
            'WithRecover': 'WithRecover() Middle',
            'WithLog': 'WithLog() Middle',
            'WithStatic': 'WithStatic(root string) Handler',
            'WithGroup': 'WithGroup(path string, handler func(g Grouper)) func(g Grouper)',
            'WithGET': 'WithGET(path string, handler Handler) func(g Grouper)',
            'WithPOST': 'WithPOST(path string, handler Handler) func(g Grouper)',
            'WithPUT': 'WithPUT(path string, handler Handler) func(g Grouper)',
            'WithDELETE': 'WithDELETE(path string, handler Handler) func(g Grouper)'
        },
        'frame': {
            'NewLogger': 'NewLogger() *log.Logger'
        }
    };

    // 导入路径补全 (import " 后提示)
    var GO_IMPORTS = [
        'fmt', 'time', 'os', 'net', 'strings', 'strconv', 'encoding/json',
        'regexp', 'errors', 'sync', 'math', 'io', 'bufio', 'path/filepath',
        'sort', 'io/ioutil', 'context', 'bytes', 'unicode', 'reflect',
        'github.com/injoyai/conv', 'github.com/injoyai/logs/v2',
        'github.com/injoyai/base/str', 'github.com/injoyai/base/crypt',
        'github.com/injoyai/base/coding', 'github.com/injoyai/base/maps',
        'github.com/injoyai/base/safe', 'github.com/injoyai/base/types',
        'github.com/injoyai/base/chans',
        // ios/v2
        'github.com/injoyai/ios/v2',
        'github.com/injoyai/ios/v2/client',
        'github.com/injoyai/ios/v2/client/dial',
        'github.com/injoyai/ios/v2/client/frame',
        'github.com/injoyai/ios/v2/client/redial',
        'github.com/injoyai/ios/v2/server',
        'github.com/injoyai/ios/v2/server/listen',
        'github.com/injoyai/ios/v2/split',
        'github.com/injoyai/ios/v2/module/common',
        'github.com/injoyai/ios/v2/module/memory',
        'github.com/injoyai/ios/v2/module/mqtt',
        'github.com/injoyai/ios/v2/module/serial',
        'github.com/injoyai/ios/v2/module/sse',
        'github.com/injoyai/ios/v2/module/ssh',
        'github.com/injoyai/ios/v2/module/tcp',
        'github.com/injoyai/ios/v2/module/udp',
        'github.com/injoyai/ios/v2/module/unix',
        'github.com/injoyai/ios/v2/module/websocket',
        // 其他 injoyai 包
        'github.com/injoyai/bar',
        'github.com/injoyai/base/crypt/aes',
        'github.com/injoyai/base/crypt/crc',
        'github.com/injoyai/base/crypt/des',
        'github.com/injoyai/base/crypt/gzip',
        'github.com/injoyai/base/crypt/md5',
        'github.com/injoyai/base/crypt/sha',
        'github.com/injoyai/base/crypt/tls',
        'github.com/injoyai/base/maps/timeout',
        'github.com/injoyai/base/maps/wait',
        'github.com/injoyai/conv/cfg',
        'github.com/injoyai/conv/codec',
        'github.com/injoyai/conv/codec/ini',
        'github.com/injoyai/conv/codec/json',
        'github.com/injoyai/conv/codec/toml',
        'github.com/injoyai/conv/codec/xml',
        'github.com/injoyai/conv/codec/yaml',
        'github.com/injoyai/frame',
        'github.com/injoyai/frame/fbr',
        'github.com/injoyai/frame/gins',
        'github.com/injoyai/frame/middle/easy_user',
        'github.com/injoyai/frame/middle/in',
        'github.com/injoyai/frame/middle/swagger',
        'i'
    ];

    function init(callback) {
        // MonacoEnvironment
        self.MonacoEnvironment = {
            getWorkerUrl: function () {
                return 'https://testingcf.jsdelivr.net/npm/monaco-editor@0.52.2/min/vs/base/worker/workerMain.js';
            }
        };
        require.config({ paths: { vs: 'https://testingcf.jsdelivr.net/npm/monaco-editor@0.52.2/min/vs' } });
        require(['vs/editor/editor.main'], function () {
            // Go 语言配置 (括号自动闭合)
            monaco.languages.setLanguageConfiguration('go', {
                autoClosingPairs: [
                    { open: '(', close: ')' },
                    { open: '[', close: ']' },
                    { open: '{', close: '}' },
                    { open: '"', close: '"', notIn: ['string'] },
                    { open: '`', close: '`', notIn: ['string'] }
                ],
                brackets: [['(', ')'], ['[', ']'], ['{', '}']]
            });

            // ===== 补全 Provider =====
            monaco.languages.registerCompletionItemProvider('go', {
                triggerCharacters: ['.', '"'],
                provideCompletionItems: function (model, position) {
                    var lineContent = model.getLineContent(position.lineNumber);
                    var before = lineContent.substring(0, position.column - 1);

                    // 1. 导入路径补全: import " 或 import 块内 "
                    var importMatch = before.match(/^\s*(?:import\s+)?"([^"]*)$/);
                    if (importMatch) {
                        var partial = importMatch[1].toLowerCase();
                        var suggestions = GO_IMPORTS.filter(function (p) {
                            return p.toLowerCase().indexOf(partial) === 0;
                        }).map(function (p) {
                            return {
                                label: p,
                                insertText: p,
                                kind: monaco.languages.CompletionItemKind.Module
                            };
                        });
                        if (suggestions.length) return { suggestions: suggestions };
                    }

                    // 2. 包成员补全: "pkg.partia"
                    var match = before.match(/(\w+)\.\s*(\w*)$/);
                    if (match) {
                        var pkg = match[1];
                        var partial = match[2].toLowerCase();
                        var suggestions = (GO_HINTS[pkg] || []).filter(function (h) {
                            return h.toLowerCase().indexOf(partial) === 0;
                        }).map(function (h) {
                            var sig = GO_SIGS[pkg] && GO_SIGS[pkg][h];
                            return {
                                label: sig || h,
                                insertText: h,
                                kind: monaco.languages.CompletionItemKind.Function,
                                detail: sig ? pkg + ' 包' : undefined
                            };
                        });
                        if (suggestions.length) return { suggestions: suggestions };
                    }

                    // 2. 通用补全: 关键字 + 内置函数 + 包名 + 文档单词 (至少2个字符)
                    match = before.match(/(\w{2,})$/);
                    if (match) {
                        var partial = match[1].toLowerCase();
                        var suggestions = [];
                        var seen = {};

                        function add(word, kind) {
                            if (word.toLowerCase().indexOf(partial) === 0 && !seen[word]) {
                                seen[word] = true;
                                suggestions.push({
                                    label: word,
                                    insertText: word,
                                    kind: kind || monaco.languages.CompletionItemKind.Keyword
                                });
                            }
                        }

                        GO_KEYWORDS.forEach(function (w) { add(w, monaco.languages.CompletionItemKind.Keyword); });
                        GO_BUILTINS.forEach(function (w) { add(w, monaco.languages.CompletionItemKind.Function); });
                        Object.keys(GO_HINTS).forEach(function (w) { add(w, monaco.languages.CompletionItemKind.Module); });

                        // 文档单词补全
                        var docText = model.getValue();
                        var words = docText.match(/\b[A-Za-z_]\w*\b/g) || [];
                        words.forEach(function (w) { add(w, monaco.languages.CompletionItemKind.Text); });

                        if (suggestions.length) return { suggestions: suggestions };
                    }

                    return { suggestions: [] };
                }
            });

            // ===== Signature Help Provider =====
            monaco.languages.registerSignatureHelpProvider('go', {
                signatureHelpTriggerCharacters: ['(', ','],
                provideSignatureHelp: function (model, position) {
                    var lineContent = model.getLineContent(position.lineNumber);
                    var before = lineContent.substring(0, position.column - 1);

                    // 向后扫描找最近的未闭合括号
                    var depth = 0, openIdx = -1;
                    for (var i = before.length - 1; i >= 0; i--) {
                        if (before[i] === ')') depth++;
                        else if (before[i] === '(') {
                            if (depth === 0) { openIdx = i; break; }
                            depth--;
                        }
                    }
                    if (openIdx < 0) return null;

                    var prefix = before.substring(0, openIdx);
                    var sig = null;

                    // 先匹配 pkg.funcName
                    var match = prefix.match(/(\w+)\.(\w+)$/);
                    if (match) {
                        sig = GO_SIGS[match[1]] && GO_SIGS[match[1]][match[2]];
                    } else {
                        // 再匹配内置函数
                        var bmatch = prefix.match(/(\w+)$/);
                        if (bmatch) {
                            sig = GO_BUILTIN_SIGS[bmatch[1]];
                        }
                    }
                    if (!sig) return null;

                    // 解析签名提取参数
                    var pStart = sig.indexOf('(');
                    var pEnd = sig.lastIndexOf(')');
                    if (pStart < 0 || pEnd < 0) return null;

                    var paramStr = sig.substring(pStart + 1, pEnd);
                    var params = paramStr ? paramStr.split(',').map(function (p) { return p.trim(); }) : [];

                    // 计算当前参数位置
                    var args = before.substring(openIdx + 1);
                    var activeParam = args.split(',').length - 1;

                    return {
                        value: {
                            signatures: [{
                                label: sig,
                                parameters: params.map(function (p) { return { label: p }; })
                            }],
                            activeSignature: 0,
                            activeParameter: Math.min(activeParam, params.length - 1)
                        },
                        dispose: function () {}
                    };
                }
            });

            // ===== 创建编辑器 =====
            var editorOpts = {
                language: 'go',
                automaticLayout: true,
                tabSize: 4,
                fontSize: 13,
                minimap: { enabled: false },
                scrollBeyondLastLine: false,
                lineNumbers: 'on',
                matchBrackets: 'always',
                autoClosingBrackets: 'always'
            };

            var addEditor = monaco.editor.create(document.getElementById('addEditor'), Object.assign({
                value: DEFAULT_SCRIPT,
                theme: 'vs-dark'
            }, editorOpts));

            var editEditor = monaco.editor.create(document.getElementById('editEditor'), Object.assign({
                value: '',
                theme: 'vs-dark'
            }, editorOpts));

            var settingEditor = monaco.editor.create(document.getElementById('settingEditor'), Object.assign({
                value: '',
                theme: 'vs-dark'
            }, editorOpts));

            if (typeof callback === 'function') {
                callback({ addEditor: addEditor, editEditor: editEditor, settingEditor: settingEditor });
            }
        });
    }

    return {
        init: init,
        DEFAULT_SCRIPT: DEFAULT_SCRIPT,
        ERROR_HANDLER_SCRIPT: ERROR_HANDLER_SCRIPT
    };
})();
