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
  #### Modelos de Entidades com unique index
  * Para usar modelos de entidade com campos de indice único você deve, além de atribuir a tag ``gorm:"uniqueIndex:{nome do indice}"`` com o nome do índice nos campos que você quer, sobrescrever também o campo `DeletedAt` incluindo a mesma tag de indice único.

  
   ``` golang
      //Exemplo com índice único:
      type Player struct {
          AuditableModel
          Name      string `gorm:"uniqueIndex:unq_name"` //Tag de indice único com a identificação do nome do índice
          NickName  string `gorm:"uniqueIndex:unq_name"` //Inclua a mesma indentificação de tag nos demais campos que fazem parte da constraint unique. 
          Address string
          DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:unq_name"` //Inclua a mesma identificação de tag no campo DeletedAt
      }
      
      
   ```