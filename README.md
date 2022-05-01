# bbbid简介

bbbid是一个高性能的分布式ID生成器，使用微服务框架kratos开发；
实现 [美团Leaf segment算法](https://tech.meituan.com/2019/03/07/open-source-project-leaf.html)，可分布式部署，保证高可用；

#### Leaf Server

![bbbid-cn](https://github.com/imsoul/bbbid/blob/dev/assets/bbbid-cn.png?raw=true)


## 环境准备

### 安装Kratos
详情请参考：https://go-kratos.dev/docs/getting-started/start

更新依赖包：
```bigquery
go get -u github.com/google/subcommands
go get -u golang.org/x/tools/go/ast/astutil
go get -u golang.org/x/tools/go/types/typeutil
go get -u golang.org/x/mod/semver
go get -u golang.org/x/tools/go/packages
```

生成所有proto源码、wire等等
```bash
go generate ./...
```

导入数据库

```sql
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for segment
-- ----------------------------
DROP TABLE IF EXISTS `segment`;
CREATE TABLE `segment`  (
  `ckey` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL COMMENT '业务key',
  `type` tinyint(0) UNSIGNED NOT NULL DEFAULT 1 COMMENT '类型：1-自增，2-随机',
  `step` int(0) UNSIGNED NOT NULL DEFAULT 0 COMMENT '步长',
  `maxid` bigint(0) UNSIGNED NOT NULL DEFAULT 0 COMMENT '最大ID',
  `intro` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '备注说明',
  `addtime` int(0) UNSIGNED NOT NULL DEFAULT 0 COMMENT '添加时间',
  `uptime` int(0) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`ckey`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_bin ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of segment
-- ----------------------------
INSERT INTO `segment` VALUES ('rand', 2, 1000, 14000, '随机ID', 1647680509, 1647680509);
INSERT INTO `segment` VALUES ('test', 1, 1000, 29248000, '自增ID', 1591706686, 1620815148);

SET FOREIGN_KEY_CHECKS = 1;
```



修改配置文件：mysql和redis配置

```yaml
server:
  http:
    addr: 0.0.0.0:8810
    timeout: 1s
  grpc:
    addr: 0.0.0.0:8811
    timeout: 1s
data:
  database:
    driver: mysql
    dsn: test:123456@tcp(127.0.0.1:3306)/bbbid?charset=utf8mb4&parseTime=True&loc=Local
    max_conns: 100
    idle_conns: 10
    life_time: 1800s
    idle_time: 600s
  redis:
    addr: 127.0.0.1:6379
    read_timeout: 0.2s
    write_timeout: 0.2s

```



运行项目

```bash
kratos run
```

请求获取ID

```
http://127.0.0.1:8810/v1/getId/test
```



添加业务

```
http://127.0.0.1:8810/v1/addBiz?ckey=demo1&type=2&step=1000&maxid=10000&intro=新业务ID
```

参数说明

| 参数名 | 类型   | 说明                               |
| ------ | ------ | ---------------------------------- |
| ckey   | string | 业务唯一KEY                        |
| type   | int    | 生成ID类型：1-有序递增，2-范围随机 |
| step   | int    | 步长                               |
| maxid  | int    | 起始最大ID                         |
| intro  | string | 业务说明                           |



#### 压测

```
wrk -c500 -t10 -d10s -T1s http://127.0.0.1:8810/v1/getId/test
```



![wrk](https://github.com/imsoul/bbbid/blob/dev/assets/wrk2.png?raw=true)



![bench](https://github.com/imsoul/bbbid/blob/dev/assets/bench2.png?raw=true)
