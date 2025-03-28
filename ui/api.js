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

