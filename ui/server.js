function connectSSE() {
    let eventSource = new EventSource("http://localhost:8080/sse/event");

    eventSource.onmessage = (event) => {
        const data = JSON.parse(event.data)
        console.log('onmessage启动', event,"new")
        for (let i = 0; i < globalData.taskData.length; i++) {
            let task = globalData.taskData[i];
            if (task.id === event.lastEventId) {
                console.log('匹配成功')
                console.log('task',task)
                task.remainingTime = data.remainingTime
                task.progress = data.progress
                task.downloadSpeed = data.downloadSpeed
                task.status = data.status
            }
        }
    }

    eventSource.onopen = () => {
        console.log('sse连接成功');
    };

    eventSource.onerror = (error) => {
        console.error(`SSE 连接错误`, error);
        eventSource.close();
    };
}

