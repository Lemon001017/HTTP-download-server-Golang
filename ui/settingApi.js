
const BASE_URL = "http://localhost:8080"
async function fetchFileList(params) {
    
    if (params.path === '/' && params.type === '' && params.sort === '' && params.order === 'down') { // 返回保存目录下的所有文件列表，默认所有类型，默认是order是down，sort 是空
        return [
            {
                name: 'Test1',
                isDirectory: true,
                path: '/data/mock_dir/test1',
                gmtModified: '2023-11-11',
                children: [
                    {
                        name: 'Test1-1',
                        isDirectory: true,
                        path: '/test1-1',
                        size: null,
                        gmtModified: '2023-11-11',
                        children: [{
                            name: 'Test1-1-1',
                            isDirectory: true,
                            path: '/test1-1-1',
                            size: '100KB',
                            gmtModified: '2023-11-11',
                            children: [{
                                name: 'Test1-1-1-1',
                                isDirectory: false,
                                type: 'txt',
                                path: '/test1-1-1-1',
                                size: '100KB',
                                gmtModified: '2023-11-11'
                            }]
                        }]
                    },
                    {
                        name: 'Test1-2.pdf',
                        isDirectory: false,
                        type: 'pdf',
                        path: '/test1-2',
                        size: '100KB',
                        gmtModified: '2023-11-11'
                    }]
            },
            {
                name: 'Test2.txt',
                isDirectory: false,
                type: 'txt',
                path: '/data/mock_dir/test2',
                size: '20KB',
                gmtModified: '2023-11-13'
            },
            {
                name: 'Test3.gif',
                isDirectory: false,
                type: 'gif',
                path: '/data/mock_dir/test3',
                size: '1MB',
                gmtModified: '2023-11-12'
            },
        ]
    }
    if (path === '//data/mock_dir/test1') {
        return [
            {
                name: 'Test1-1',
                isDirectory: true,
                path: '/test1-1',
                size: null,
                gmtModified: '2023-11-11',
                children: [{
                    name: 'Test1-1-1',
                    isDirectory: true,
                    path: '/test1-1-1',
                    size: '100KB',
                    gmtModified: '2023-11-11',
                    children: [{
                        name: 'Test1-1-1-1',
                        isDirectory: false,
                        type: 'txt',
                        path: '/test1-1-1-1',
                        size: '100KB',
                        gmtModified: '2023-11-11'
                    }]
                }]
            },
            {
                name: 'Test1-2.pdf',
                isDirectory: false,
                type: 'pdf',
                path: '/test1-2',
                size: '100KB',
                gmtModified: '2023-11-11'
            }]
    }
    if (path === '//data/mock_dir/test2') {
        return [
            {
                name: 'Test2.txt',
                isDirectory: false,
                type: 'txt',
                path: '/data/mock_dir/test2',
                size: '20KB',
                gmtModified: '2023-11-13'
            }
        ]
    }
    if (path === '/test1-1') {
        return [
            {
                name: 'Test1-1-1',
                isDirectory: true,
                path: '/test1-1-1',
                size: '100KB',
                gmtModified: '2023-11-11',
                children: [{
                    name: 'Test1-1-1-1',
                    isDirectory: false,
                    type: 'txt',
                    path: '/test1-1-1-1',
                    size: '100KB',
                    gmtModified: '2023-11-11'
                }]
            }
        ]
    }
    else {
        const resp = await fetch(BASE_URL + "/api/file/list", {
            method: "POST",
            headers: {
                "Content-Type":"application/json"
            },
            body:JSON.stringify({"path":params.path,"type":params.type,"sort":params.sort,"order":params.order})
        })
       // 
        const data = await resp.json()
       // 
        return data["data"];
    }

}
