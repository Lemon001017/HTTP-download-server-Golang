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
// async function fetchFileList(params,fileName) {
//     const resp = await fetch(BASE_URL + "/api/file/list?fileName="+fileName, {
//         method: "POST",
//         headers: {
//             "Content-Type":"application/json"
//         },
//         body:JSON.stringify({"type":params.type,"sort":params.sort,"order":params.order})
//     })
//     const data = await resp.json()
//     console.log('data:',data);
//     return data.data;
// }

