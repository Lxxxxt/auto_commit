第一步： 开通 Doubao-pro-32k 链接：https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement?OpenTokenDrawer=false
![image](https://github.com/user-attachments/assets/28c72a03-4665-4065-9e92-7f641eed8052)

第二步： 获取api key
![image](https://github.com/user-attachments/assets/c7eb5afd-792c-46f0-820e-b22cfbff7e57)

第三步：安装auto_commit
  go install github.com/Lxxxxt/auto_commit 

第四步： 配置别名
  alias autoc='auto_commit'
  
第五步： 设置api key
  autoc -k {{your_api_key}}
