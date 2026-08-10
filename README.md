PS C:\Users\CORE5 ULTRA\OneDrive\Desktop\proyecto institucional\microservicio\design-software-iam-db-back\iam-service> tree /F /A
Listado de rutas de carpetas para el volumen Windows
El número de serie del volumen es 000000F3 0680:5D85
C:.
|   .env.develop
|   .env.main
|   .env.qa
|   .env.staging
|   cookies.txt
|   docker-compose.yml
|   Dockerfile
|   go.mod
|   go.sum
|   login.json
|   tmp_hash.go
|   
+---cmd
|   +---api
|   |       main.go
|   |       
|   \---seed
|           main.go
|           
+---internal
|   +---api
|   |   +---dto
|   |   |       auth_dto.go
|   |   |       catalog_dto.go
|   |   |       user_dto.go
|   |   |       
|   |   \---http
|   |       |   router.go
|   |       |   
|   |       +---handlers
|   |       |       auth_handler.go
|   |       |       feature_handler.go
|   |       |       module_handler.go
|   |       |       rbac_handler.go
|   |       |       user_handler.go
|   |       |       
|   |       \---middleware
|   |               auth_middleware.go
|   |               cors.go
|   |               
|   +---application
|   |   +---auth
|   |   |       dto.go
|   |   |       forgot_password_usecase.go
|   |   |       login_usecase.go
|   |   |       logout_usecase.go
|   |   |       refresh_token_usecase.go
|   |   |       reset_password_usecase.go
|   |   |       
|   |   +---catalog
|   |   |       create_feature_usecase.go
|   |   |       create_module_usecase.go
|   |   |       list_features_by_module_usecase.go
|   |   |       list_features_usecase.go
|   |   |       list_modules_usecase.go
|   |   |       update_feature_usecase.go
|   |   |       update_module_usecase.go
|   |   |       
|   |   +---rbac
|   |   |       assign_role_usecase.go
|   |   |       
|   |   \---user
|   |           create_user_usecase.go
|   |           dto.go
|   |           get_user_usecase.go
|   |           list_users_usecase.go
|   |           set_user_status_usecase.go
|   |           unlock_user_usecase.go
|   |           update_user_usecase.go
|   |           
|   +---domain
|   |   +---catalog
|   |   |       feature.go
|   |   |       module.go
|   |   |       repository.go
|   |   |       
|   |   +---passwordreset
|   |   |       entity.go
|   |   |       repository.go
|   |   |       
|   |   +---role
|   |   |       entity.go
|   |   |       repository.go
|   |   |       
|   |   +---session
|   |   |       entity.go
|   |   |       repository.go
|   |   |       
|   |   \---user
|   |           entity.go
|   |           errors.go
|   |           repository.go
|   |           
|   \---infrastructure
|       +---messaging
|       +---notification
|       |       smtp_sender.go
|       |       
|       +---persistence
|       |   \---postgres
|       |           audit_repository.go
|       |           db.go
|       |           feature_repository.go
|       |           module_repository.go
|       |           password_reset_repository.go
|       |           role_repository.go
|       |           session_repository.go
|       |           user_repository.go
|       |           
|       \---security
|               jwt.go
|               password_hasher.go
|               
+---pkg
|   \---config
|           config.go
|           
\---shared


$body = @{
    email = "admin@sena.edu.co"
    password = "Admin12345!"
} | ConvertTo-Json

Invoke-RestMethod `
    -Method POST `
    -Uri "http://localhost:8001/auth/login" `
    -ContentType "application/json" `
    -Body $body

    package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	user = "admin@sena.edu.co"
	password := "Admin12345!"

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(hash))
}


docker compose --env-file .env.develop down
docker compose --env-file .env.develop up -d --build
docker compose --env-file .env.develop config <-Verifica que Docker esté recibiendo las variables
docker compose --env-file .env.develop ps <- comprobar que el contenedor esta levantado
docker ps -a <- ver todo lo contenedores



 $loginBody = @{
>> 
>>   email    = "admin@sena.edu.co"
>> 
>>   password = "Admin12345!"
>> 
>> } | ConvertTo-Json
>> 
>> $loginResponse = Invoke-RestMethod -Uri "http://localhost:8001/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
>> 
>> $loginResponse
>> 
>> $token = $loginResponse.access_token
>> 
>> $headers = @{ Authorization = "Bearer $token" }