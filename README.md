# minitan
a restful http service in go

## there are two databases named relationship and users

the struct of database relationship

"mainuser"	"touser"	"liked"	"is_matched"	"no"

 ubigint     ubigint  utinyint  utinyint   utinyint
 
When two users liked each other ,both the "liked"	"is_matched" are set "1"

Once a user "dislike" the other,"liked"=0	"is_matched"=0 in the row of the user are set.And "is_matched"=0 in the row of the other user is set


the struct of database users

"id"	  "name"

ubigint string

id is an  Auto Increment primary key

#  environment
postgreSQL 13 (driver:	_ "github.com/lib/pq")

go version go1.15.6 darwin/amd64

route driver:"github.com/gorilla/mux"

Goland 2020.3

# pics
![list all users](https://github.com/YiTongG/minitan/blob/main/1.png)
![create users](https://github.com/YiTongG/minitan/blob/main/2.png)
![list all users relationship](https://github.com/YiTongG/minitan/blob/main/3.png)
![update users relationship](https://github.com/YiTongG/minitan/blob/main/4.png)


