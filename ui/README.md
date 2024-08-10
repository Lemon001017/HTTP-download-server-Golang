# Ruzhila project reference UI

ps: 大家的注释说明都要用英文，这里用中文是为了方便大家理解，后期会统一修改为英文

## How to use the UI
- 在spring工程的src/main/resources下创建public目录
- 将本工程的所有文件放在public目录下
- 启动springboot工程，访问即可

- Create a public directory in the src/main/resources directory of the spring project
- Put all the files of the project in the public directory
- Start the SpringBoot project and access it

## The technology used by the UI
- [alpinejs.dev](https://alpinejs.dev/) // Front-end framework
- [tailwindcss.com](https://tailwindcss.com/) // css framework
- [heroicons.com](https://heroicons.com/) // Icon Library
- [tailwindui.com](https://tailwindui.com/) // Component Library

## 关于接口说明
- 接口位置：src/main/resources/public/js/api.js
- 接口说明：接口返回的数据格式为json，其中code为状态码，data为返回的数据
- 接口的具体返回需要的数据，格式，统一在api.js中进行说明


- index.html 是文档列表页面，文档列表页面需要调用接口获取文档列表数据
- setting.html 是设置页面，设置页面需要调用接口获取设置数据，可以修改设置数据，点击保存，保存数据
- tasks.html 是任务列表页面，任务列表页面需要调用接口获取任务列表数据，这里面有更详细的操作，过滤，翻页，单量、批量操作删除，刷新，暂停任务下载等功能
