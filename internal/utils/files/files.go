package files

import (
	"fmt"
	"os"
)

/* Verifica apenas se o arquivo existe. */
func FileExist(fullFileName string) bool {
	_, err := os.Stat(fullFileName)
	return !os.IsNotExist(err)

}

// Deleta um arquivo
func DeletarFile(fullFileName string) error {
	err := os.Remove(fullFileName)
	if err != nil {
		fmt.Printf("Erro ao deletar o arquivo: %s\n", err)
		return err
	}
	fmt.Printf("Arquivo \"%s\" deletado com sucesso.\n", fullFileName)
	return nil
}
