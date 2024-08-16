# serverGoChi
## Gen database models form mysql database

go install gorm.io/gen/tools/gentool@latest

gentool -dsn "`user`:`pw`@tcp(`ip`:`port`)/`database`?charset=utf8mb4&parseTime=True&loc=Local" 

