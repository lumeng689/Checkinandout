### Build
0. Please use Go version `1.14` for building. (Tested on `1.14.15`)
1. In cmd, run `go build`
2. To build the portal, go the the `/web` Directory and run `npm install; npm build`

### Run
0. Create an empty `data` folder under project root
1. click `startup.bat` and let it run
2. A portal website will be hosted at `localhost:8000\#\cc-records`. Please open the url in browser to use it
##### Alternatively
1. Please startup a mongodb server. For local mongo server, in cmd run `mongod --dbpath=/data` (you can replace the "/data" with any directory you wish; you need to create the folder before running `mongod`). If cmd reports "mongod not recognized", please install mongodb or check environment variables.
2. In cmd, run `.\cc-server.exe`
3. A portal website will be hosted at `localhost:8000\#\cc-records`. Please open the url in browser to use it

### Update Twilio SMS Token:
1. Login to `mongo` shell, and switch db by `use go_mongo`
2. Update the Token using the following shell command:
`> db.configs.update({name: "default"}, {$set: {sms_auth_token: "65a64755b83eb1e8e6ece4a7e7b6bce7"}})` 
If Succeed, the shell returns info: `WriteResult({ "nMatched" : 1, "nUpserted" : 0, "nModified" : 1 })`