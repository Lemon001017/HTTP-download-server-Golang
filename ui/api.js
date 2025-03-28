const BASE_URL = "http://localhost:8000"
// get settings
async function fetchSettings() {
    const response = await fetch(`${BASE_URL}/api/settings/1`);
    return (await response.json());
}


// save settings
async function saveSettings(params) {
    const resp = await fetch(`${BASE_URL}/api/settings/1`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(params)
    })
    const setting = await resp.json();
    return setting;
}

// get file list
async function fetchFileList(params, path = "") {
    const resp = await fetch(`${BASE_URL}/api/file/list`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            "path": params.path || path,
            "type": params.type,
            "sort": params.sort,
            "order": params.order
        })
    })
    const data = await resp.json()
    return data.data;
}

// rename file
async function renameFile(path, newName) {
    const resp = await fetch(`${BASE_URL}/api/file/rename`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            "path": path,
            "newName": newName
        })
    })
    const data = await resp.json()
    return data;
}

// delete file
async function deleteFile(path) {
    const resp = await fetch(`${BASE_URL}/api/file/delete`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            "path": path
        })
    })
    const data = await resp.json()
    return data;
}

// create directory
async function createDirectory(path, dirName) {
    const resp = await fetch(`${BASE_URL}/api/file/mkdir`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            "path": path,
            "dirName": dirName
        })
    })
    const data = await resp.json()
    return data;
}

// search files
async function searchFiles(path, query, exact = false, recursive = true) {
    const resp = await fetch(`${BASE_URL}/api/file/search`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            "path": path,
            "query": query,
            "exact": exact,
            "recursive": recursive
        })
    })
    const data = await resp.json()
    return data.data;
}

