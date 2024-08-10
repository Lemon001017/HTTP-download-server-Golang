// const BASE_URL = "http://118.25.40.30:8081"
const BASE_URL = "http://localhost:8080"
// 设置页面 api
async function fetchSettings() {
    const response = await fetch(`${BASE_URL}/api/settings/get`);
    return (await response.json())["data"];
}


// 保存设置
async function saveSettings(params) {
    const resp = await fetch(`${BASE_URL}/api/settings/save`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(params)
    })
    const setting = await resp.json();
    return setting["data"];
}

// file 页面 api
async function fetchFileList(params,fileName) {
    const resp = await fetch(BASE_URL + "/api/file/list?fileName="+fileName, {
        method: "POST",
        headers: {
            "Content-Type":"application/json"
        },
        body:JSON.stringify({"type":params.type,"sort":params.sort,"order":params.order})
    })
    //
    const data = await resp.json()
    console.log('data:',data);
    return data.data;
}

