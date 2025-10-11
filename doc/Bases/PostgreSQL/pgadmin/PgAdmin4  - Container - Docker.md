# PgAdmin4 - Usando Container

O uso do PgAdmin4 por meio de uma instalação local ficou muito complicado, pois ele só está disponível
por emio do snap e esse instalador cria uma série de restrições de segurança que impedem o uso das fer-
ramentas de backup e restore. Ele alerta qeu o pg_restore, pg_backup e etc não estão disponíveis no
/usr/bin, mesmo estando. Fiz muitas tentativas até descobrir que era o snap. Para continuar usando o 
snap eu teria que escancará a segurança e preferi usa container de pgadmin4 para as intervenções. o
comando mais simples é o abaixo. 

$ docker run --name pgadmin -p 5050:80 -e PGADMIN_DEFAULT_EMAIL=admin@admin.com -e PGADMIN_DEFAULT_PASSWORD=admin -d dpage/pgadmin4

Depois, acesse no navegador:
➡ http://localhost:5050
Login: admin@admin.com
Senha: admin

# backup e restore

Descobri que a janela de restore de backups permite o upload e seleção de arquivos dentro do cliente, 
sem precisar de uma ginástica para copiar o arquivo para o container e depois restaurar.

