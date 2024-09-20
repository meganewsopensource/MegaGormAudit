# Mega Gorm Audit

## O que é?

Plugin para o [Gorm](https://gorm.io/index.html) que adiciona recurso de auditoria de registros.

---
## Instalação
```shell
go get github.com/meganewsopensource/megagormaudit
```
---
## Uso

### Setup do gorm
```golang
    //criação do objeto DB
    db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}))

	//Uso do plugin
	err = db.Use(MegaGormAuditPlugin{})
```

### Modelos de Entidades

* Sempre que um novo estado é atribuido a um registro,  um novo registro é inserido com o novo estado e o anterior é marcado logicamente como "removido".
Para isso, atribua a struct `AuditableModel` em seu modelo de banco de dados.
    ```golang
     type Company struct {
        AuditableModel //atribução do modelo de dados auditável
        //demais propriedades
        Name string
        Address string
    }
    ```
    Ao atribuir `AuditableModel` as seguintes propriedades serão adicionadas à struct:
    ```golang
      ID             //chave primária da tabela autoincremtada     
      AuditParentID  //chave estrangeira que liga o registro pai da auditoria
      AuditParent    //representação do objeto AuditableModel pai da auditoria
      CreatedAt      //data e hora da criação do registro
      UpdatedAt       //data e hora da atualização do registro
      DeletedAt      //data e hora de deleção lógica do registro. Flag para atribuir a deleção lógica
      LastChangedUser //identificação do usuário que fez a ulima alteração dos dados.
    ```
