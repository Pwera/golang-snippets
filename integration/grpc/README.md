```
grpcurl --plaintext 0.0.0.0:10000 list
grpcurl --plaintext 0.0.0.0:10000 describe protocol.User
grpcurl --plaintext 0.0.0.0:10000 describe protocol.NewUserRequest
grpcurl -plaintext -format text -d 'email: "email@email.com"'   localhost:10000 protocol.User.SubmitNewUser

```