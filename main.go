package main

import (
	"context"
	"flag"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	sourceSecretName      = flag.String("source", "", "Name of the source secret")
	sourceSecretNamespace = flag.String("source-ns", "default", "Namespace of the source secret")
	destSecretName        = flag.String("dest", "", "Name of the destination secret")
	destSecretNamespace   = flag.String("dest-ns", "default", "Namespace of the destination secret")
)

func main() {
	flag.Parse()

	cl := initClient()

	if *destSecretName == "" {
		*destSecretName = *sourceSecretName
	}

	secret, err := cl.CoreV1().Secrets(*sourceSecretNamespace).Get(context.Background(), *sourceSecretName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	copiedData := copyMap(secret.Data)
	dstSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: *destSecretName, Namespace: *destSecretNamespace},
		Data:       copiedData,
	}

	_, err = cl.CoreV1().Secrets(*destSecretNamespace).Get(context.Background(), *destSecretName, metav1.GetOptions{})
	if err == nil {
		_, err = cl.CoreV1().Secrets(*destSecretNamespace).Update(context.Background(), &dstSecret, metav1.UpdateOptions{})
		if err != nil {
			panic(err)
		}

		fmt.Printf("Secret is updated!\n%v/%v\n", dstSecret.Name, dstSecret.Namespace)

		return
	}

	// Instead of creating the secret directly, use CreateOrUpdate
	_, err = cl.CoreV1().Secrets(*destSecretNamespace).Create(context.Background(), &dstSecret, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Secret is created!\n%v/%v\n", dstSecret.Name, dstSecret.Namespace)
}

func initClient() *kubernetes.Clientset {
	clientset := kubernetes.NewForConfigOrDie(config.GetConfigOrDie())

	return clientset
}

func copyMap(original map[string][]byte) map[string][]byte {
	// Create a new map to hold the copied data
	copiedMap := make(map[string][]byte)

	// Iterate over the original map
	for key, value := range original {
		// Create a new byte slice to store the copied value
		copiedValue := make([]byte, len(value))
		copy(copiedValue, value)

		// Add the key-value pair to the copied map
		copiedMap[key] = copiedValue
	}

	return copiedMap
}
