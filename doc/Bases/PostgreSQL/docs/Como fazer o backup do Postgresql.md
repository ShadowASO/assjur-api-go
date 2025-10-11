PROCEDIMENTO PARA REALIZAR BACKUP DO DATABASE NO POSTGRESQL

1) Acessar o database utilizando PgAdmin, utilizando o container;
2) Selecionar o Database desejado(ex. assjurdb);
3) Clica no botão esquerdo do mause e seleciona a opção "backup" no meu pop-up;
4) Janela Backup
    . Insere o nome para o arquivo do backup;
    . Format: muda o formato para "Plain";
    . Query Options: selecione "Include CREATE DATABASE statement";
    . Clica botão Backup;
    . O backup será gerado dentro do container do PgAdmin que roda na máquina do usuário local;
    
5) Copia o arquivo do backup do container para a pasta desejada no host:
    . docker cp pgadmin:/var/lib/pgadmin/storage/admin_admin.com/bkp-home-03082025-2.sql .
    . docker cp <ID CONTAINER>:/var/lib/pgadmin/storage/admin_admin.com/<nome do arquivo> <destino>




