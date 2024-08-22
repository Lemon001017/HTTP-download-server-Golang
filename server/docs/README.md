### 分片下载实现步骤

1. 用户提交下载 url
2. 获取文件大小、保存路径、最大速度等
3. 计算分片大小、分片数量，初始化所有分片
4. 创建 goroutine 池，循环将所有分片丢入池中
5. 动态计算下载速度、进度、剩余时间，并通过 sse 发送到客户端
6. 等待下载结束

### 暂停下载实现步骤

1. 获取所有 id
2. 根据 id 获取所有任务
3. 枚举所有任务，更新任务状态，然后调用 cleanEventSource 方法执行上下文取消

### 恢复下载实现步骤

`关键：每个下载任务都对应一个记分牌，用于记录 taskId、分片索引、开始位置、结束位置和是否下载完成，这些数据存入 chunks 表中；
   当这个分片下载完成就会标记为 true 并更新数据库，这样再次开启下载只需要提交未下载的分片到 pool 中即可。`

1. 获取所有 id
2. 根据 id 获取所有任务
3. 枚举所有任务，更新任务状态，根据任务 id 获取所有分片，异步调用 processDownload 方法执行下载
4. 等待下载结束

### 重新下载实现步骤

1. 获取所有 id
2. 根据 id 获取所有任务
3. 枚举所有任务，根据 key 创建对应的 eventSource，更新任务状态，异步调用 processDownload 方法执行下载
4. 等待下载结束

### 限速下载实现步骤

1. 安装 `golang.org/x/time/rate` 库
2. 创建一个限速器，设置每秒允许的请求数和桶的大小(这里限速单位是 MB/s，需要先将 MB/s 转换为 Bytes/s)

```go
	maxDownloadSpeedInBytes := maxDownloadSpeed * 1000 * 1000
	limiter := rate.NewLimiter(rate.Limit(maxDownloadSpeedInBytes), int(maxDownloadSpeedInBytes))
```

3. 集成到 `downloadChunk`方法中
