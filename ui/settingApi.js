const BASE_URL = "http://localhost:8000"
async function fetchFileList(params) {
    const resp = await fetch(BASE_URL + "/api/file/list", {
        method: "POST",
        headers: {
            "Content-Type":"application/json"
        },
        body:JSON.stringify({"path":params.path,"type":params.type,"sort":params.sort,"order":params.order})
    })
    const data = await resp.json()
    return data["data"];
}
