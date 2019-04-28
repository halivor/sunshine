# tcp connection
```
 消息内容 "message" , 接收Ack消息
 单核CPU, > 35000连接, 每秒 4条消息, CPU达到100%, PPS >  28万
 单核CPU, >  8000连接, 每秒20条消息, CPU达到100%, PPS ~= 32万
```
```
Header  // 内部通讯格式，与外部无关

SHeader // 与客户端交互格式，字符串头
> ver   "10"
> typ   'B' 二进制头
>      'S' 字符串头
> opt   '0' 普通消息
>      '1' 消息需要ack
>      'A' ack消息
> cmd       命令号 100, typ = 'S', cmd = '0100' 
>                       typ = 'B', cmd =   100 , 使用大端表示
> seq       待确认消息sequence
> len       长度, typ = 'S' => "0123" 代表sheader后有123个字符
>                 typ = 'B' =>   123 代表sheader后有123个字符, 使用大端表示
> res       预留


Auth
> ver
> typ
> opt
> uid       连接唯一ID, uid:10000 => "000100000"
> cid       连接分类ID, rid:10000 => "000100000"
> sign      签名串
> len       Auth包后字符串长度
```
