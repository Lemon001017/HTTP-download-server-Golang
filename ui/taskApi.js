// const BASE_URL = "http://118.25.40.30:8081"
const BASE_URL = "http://localhost:8080"
async function fetchTasks(params) {

    const resp = await fetch(`${BASE_URL}/api/task/get_tasks`, {
        method: "POST",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded"
        },
        body: `currentPage=${params.currentPage}&pageSize=${params.limit}&filter=${params.status}`

    })
    const result = await resp.json()

    return {
        total: result?.data?.totalCount,
        items: result?.data?.data
    }
}


// 对任务的状态进行过滤选择,如果是all 的情况下，就返回所有的数据，默认是all 的情况
async function fetchFilterTasks(filter, pos, limit, currentPage) {

    let total = 0
    let items = []
    if (filter === 'all') {
        await fetch(BASE_URL + "/api/task/get_tasks", {
            method: "POST",
            headers: {
                "Content-Type": "application/x-www-form-urlencoded"
            },
            body: `currentPage=${currentPage}&pageSize=${limit}&filter=${filter}`
        }).then(data => {
            return data.json()
        }).then(response => {
            if (response.code === 200) {
                total = response.data.total
                items = response.data.items
            }
        })
        return {
            total: total,
            items: items
        }
    } else {
        await fetch(BASE_URL + "/api/task/get_tasks", {
            method: "POST",
            headers: {
                "Content-Type": "application/x-www-form-urlencoded"
            },
            body: `currentPage=${currentPage}&pageSize=${limit}&filter=${filter}`
        }).then(data => {
            return data.json()
        }).then(response => {
            if (response.code === 200) {
                total = response.data.total
                items = response.data.items
            }
        })
        return {
            total: total,
            items: items
        }
    }
}

async function changeThreads(params) {

    let isSuccess = false
    const resp = await fetch(BASE_URL + "/api/task/update_thread", {
        method: "POST",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded"
        },
        body: `id=${params.id}&threads=${params.threads}`
    })
    const data = await resp.json()

    if (data.code === 200) {
        isSuccess = true
    } else {
        isSuccess = false

    }
    if (isSuccess) {
        return data["data"];
    } else {
        return {}
    }
}

async function submitDownloadPath(path) {
    const resp = await fetch(BASE_URL + "/api/task/submit", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({"url": path})
    })
    const data = await resp.json()
    //connectSSE()
    return data.data;
}

// 重新下载任务的详细信息，ids是一个数组，单个任务，就是一个元素的数组，多个任务就是多个元素的数组，实现同一个接口单量和多量的处理
async function fetchTaskInfo(ids) {

    let isRefresh = false
    const resp = await fetch(BASE_URL + "/api/task/refresh", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(ids)
    })
    const data = await resp.json()

    if (resp.code === 200) {
        isRefresh = true
        location.reload()
    }
    //return isRefresh
    return data["data"];
}

// 暂停下载任务,ids是一个数组，单个任务，就是一个元素的数组，多个任务就是多个元素的数组，实现同一个接口单量和多量的处理
async function pauseTask(ids) {
    const resp = await fetch(BASE_URL + "/api/task/pause", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(ids)
    })
    const data = await resp.json()
    return data["data"];
}

// 恢复下载任务，ids是一个数组
async function resumeTask(ids) {
    const resp = await fetch(BASE_URL + "/api/task/resume", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(ids)
    })
    const data = await resp.json()
    return data["data"];
}

// 删除下载任务，ids是一个数组
async function deleteTask(ids) {
    const resp = await fetch(BASE_URL + "/api/task/delete", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(ids)
    })
    const data = await resp.json()
    return data["data"];
}