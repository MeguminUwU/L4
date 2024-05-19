package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "math/rand"
    "os"
    "strconv"
    "sync"
    "time"
    "strings"

    pb "L4/proto"

    "google.golang.org/grpc"
)

func sendMessage(client pb.MercServiceClient, id, accion string) (*pb.MercResponseMessage, error) {
    ctx := context.Background() 

    req := &pb.MercRequestMessage{
        ID:     id,
        Accion: accion,
    }

    return client.MyMethod(ctx, req)
}

func mercBot(client pb.MercServiceClient, id string) {
    response, err := sendMessage(client, id, "Ready") // los bots comunican estar listos al inicializarse
    if err != nil {
        log.Printf("Error en bot %s al enviar mensaje inicial: %v", id, err)
        return
    }

    for {
        if response.Informacion == "Waiting" { // Si se está esperando mi confirmación para entrar al siguiente piso
            response, err = sendMessage(client, id, "Ready")
            if err != nil {
                log.Printf("Error en bot %s al enviar 'Ready': %v", id, err)
                return
            }
            continue
        } else if response.Informacion == "1" { // lógica piso 1
            weapon := strconv.Itoa(rand.Intn(3))
            response, err = sendMessage(client, id, weapon)
            if err != nil {
                log.Printf("Error en bot %s al enviar 'weapon': %v", id, err)
                return
            }
            continue
        } else if response.Informacion == "2" { // lógica piso 2
            hallway := strconv.Itoa(rand.Intn(2))
            response, err = sendMessage(client, id, hallway)
            if err != nil {
                log.Printf("Error en bot %s al enviar 'hallway': %v", id, err)
                return
            }
            continue
        } else if response.Informacion == "3" { // lógica piso 3, responder por la ronda.
            guess := strconv.Itoa(1 + rand.Intn(15))
            response, err = sendMessage(client, id, guess)
            if err != nil {
                log.Printf("Error en bot %s al enviar 'guess': %v", id, err)
                return
            }
            continue
        } else if response.Informacion == "Dead" || response.Informacion == "Win" { // Si me muero o gano, salgo del while y se finaliza el subproceso, al bot le da lo mismo si gana lo importante es
            break // queda registrado igual para que player lo vea
        }
    }
}

func main() {
    rand.Seed(time.Now().UnixNano())
    conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("fallo al conectar: %v", err)
    }
    defer conn.Close()

    client := pb.NewMercServiceClient(conn)

    var wg sync.WaitGroup

    for id := 2; id <= 8; id++ { // se inicializan los bot desde el 2 al 8 siendo el 1 reservado para el player
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            mercBot(client, strconv.Itoa(id))
            
        }(id)
    }

    // acá va la interfaz del player por consola
    reader := bufio.NewReader(os.Stdin)
    id := "1"
    response, err := sendMessage(client, id, "Ready") 

    for {
        if response.Informacion == "Waiting" { // Si se está esperando mi confirmación para entrar al siguiente piso
            fmt.Println("Confirmación: Ingrese cualquier tecla para enviar 'Ready' y continuar.")
            _ , _ = reader.ReadString('\n')
            response, err = sendMessage(client, id, "Ready")
            continue
        } else if response.Informacion == "1" { // lógica piso 1
            fmt.Println("Piso 1: Ingrese un número para seleccionar un arma (0, 1, 2):")
            weapon, _ := reader.ReadString('\n')
            weapon = strings.TrimSpace(weapon)
            response, err = sendMessage(client, id, weapon)

            continue
        } else if response.Informacion == "2" { // lógica piso 2
            fmt.Println("Piso 2: Ingrese un número para seleccionar un pasillo (0, 1):")
            hallway, err := reader.ReadString('\n')
            if err != nil {
                log.Printf("Error leyendo input: %v", err)
                return
            }
            hallway = strings.TrimSpace(hallway)
            response, err = sendMessage(client, id, hallway)

            continue
        } else if response.Informacion == "3" { // lógica piso 3
            fmt.Println("Piso 3: Ingrese un número para seleccionar un pasillo (1-15):")
            guess, _ := reader.ReadString('\n')

            guess = strings.TrimSpace(guess)
            response, err = sendMessage(client, id, guess)

            continue
        } else if response.Informacion == "Dead" || response.Informacion == "Win" { // Si me muero o gano
            fmt.Printf("Has muerto :c\n")
            break
        }
    }

    wg.Wait()
    fmt.Println("Todos los bots han finalizado")
}