package main

import (
    "context"
    "fmt"
    "log"
    "math/rand"
    "net"
    "os"
    "strconv"
    "sync"
    "bufio"

    pb "L4/proto" // Importa el paquete generado

    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

// Server es el servidor gRPC
type Server struct {
    pb.UnimplementedMercServiceServer
    mu         sync.Mutex
    reqCount   int
    reqCh      chan requestContext
}

type requestContext struct { //estructura que conserva referencia a una request
    req *pb.MercRequestMessage
    ctx context.Context
    res chan *pb.MercResponseMessage
}

var requiredRequests = 8 //valor de request requeridos antes de resolver respuestas
var Stage = 0
var Piso3Status = [3]int{0, -1, 0} //arreglo que guarda los aciertos del piso 3, la posicion 0 corresponde a un contador de la ronda actual, la 1 es la decision del patriarca y el 3 es la suma de aciertos totales

func (s *Server) MyMethod(ctx context.Context, req *pb.MercRequestMessage) (*pb.MercResponseMessage, error) {
    resCh := make(chan *pb.MercResponseMessage)
    log.Printf("Recived request: ID=%s, accion=%s", req.ID, req.Accion)
    s.mu.Lock()
    if req.Accion == "consulta" { //si la accion es consulta no cuenta para el counter
    } else {
        s.reqCount++
        s.reqCh <- requestContext{req: req, ctx: ctx, res: resCh}
        if s.reqCount == requiredRequests {
            s.processRequests(Stage) //se procesan las request al llegar al recopilar todas.
            if Stage < 5 {
                Stage++
            }
            s.reqCount = 0
        }
    }
    s.mu.Unlock()

    response := <-resCh
    return response, nil
}

func (s *Server) processRequests(stage int) {
    go func() {
        s.mu.Lock()
        defer s.mu.Unlock()
        reader := bufio.NewReader(os.Stdin)
        fmt.Printf("\n")
        initialRequiredRequests:=requiredRequests
        if stage == 0 { //confirmacion listo para mision
            log.Printf("Se han confirmado todos los participantes de la mision!")
            fmt.Printf("Si desea comenzar la mision ingrese cualquier input: ")
            _, _ = reader.ReadString('\n') // Ignora la entrada del usuario
            for i := 0; i < initialRequiredRequests; i++ {
                reqCtx := <-s.reqCh
                response := &pb.MercResponseMessage{
                    Informacion: "1",
                }
                reqCtx.res <- response
            }
            log.Printf("La Mision ha comenzado!, se le ha notificado a todos los mercenarios, una vez recibidas sus decisiones se resolvera el primer piso")
        } else if stage == 1 { //Piso 1
            log.Printf("Se han recibido todas las acciones de los mercenarios")
            X := rand.Intn(101)
            Y := rand.Intn(101)
            if Y < X {
                aux := X
                X = Y
                Y = aux
            }
            var WeaponsChances = [3]int{X, Y - X, 100 - Y}
            for i := 0; i < initialRequiredRequests; i++ {
                reqCtx := <-s.reqCh
                log.Printf("Procesando request: Mercenario %s, accion %s", reqCtx.req.ID, reqCtx.req.Accion)
                result := rand.Intn(101)
                num, _ := strconv.Atoi(reqCtx.req.Accion)
                response := &pb.MercResponseMessage{
                    Informacion: "Waiting",
                }
                if result <= WeaponsChances[num] { //Mercenario sobrevive
                    log.Printf("El Mercenario %s ha sobrevivido al piso 1 obteniendo %d/100 con chance %d", reqCtx.req.ID, result, WeaponsChances[num])
                } else {
                    log.Printf("El Mercenario %s ha muerto en el piso 1 obteniendo %d/100 con chance %d", reqCtx.req.ID, result, WeaponsChances[num])
                    response = &pb.MercResponseMessage{
                        Informacion: "Dead",
                    }
                    requiredRequests--
                }
                reqCtx.res <- response
            }
        } else if stage == 2 { //Confirmacion entrada piso 2
            log.Printf("Se han confirmado todos los participantes para continuar al siguiente piso")
            fmt.Printf("Si desea comenzar el piso 2 ingrese cualquier input: ")
            _, _ = reader.ReadString('\n') // Ignora la entrada del usuario
            for i := 0; i < initialRequiredRequests; i++ {
                reqCtx := <-s.reqCh
                response := &pb.MercResponseMessage{
                    Informacion: "2",
                }
                reqCtx.res <- response
            }
            log.Printf("El segundo piso comienza!, se le ha notificado a todos los mercenarios, una vez recibidas sus decisiones se resolvera el segundo piso")
        } else if stage == 3 { //Piso 2
            log.Printf("Se han recibido todas las acciones de los mercenarios")
            CorrectHallway := rand.Intn(2)
            CorrectHallwayS := strconv.Itoa(CorrectHallway)
            log.Printf("Se ha decidido el camino correcto!, siendo este el camino %d", CorrectHallway)
            for i := 0; i < initialRequiredRequests; i++ {
                reqCtx := <-s.reqCh
                num, err := strconv.Atoi(reqCtx.req.Accion)
                if err != nil {
                    log.Printf("error al transformar el input: %v", err)
                }
                log.Printf("Procesando request: Mercenario %s, accion %s", reqCtx.req.ID, reqCtx.req.Accion)
                response := &pb.MercResponseMessage{
                    Informacion: "Waiting",
                }
                if reqCtx.req.ID=="1"{
                    log.Printf("TU EL MERCENARIO 1, HAS ELEGIDO %d,%s", num,reqCtx.req.Accion)
                }
                if CorrectHallwayS == reqCtx.req.Accion { //Mercenario sobrevive
                    log.Printf("El Mercenario %s ha sobrevivido al piso 2 eligiendo el camino %d", reqCtx.req.ID, num)
                } else {
                    log.Printf("El Mercenario %s ha muerto en el piso 2 eligiendo el camino %d", reqCtx.req.ID, num)
                    response = &pb.MercResponseMessage{
                        Informacion: "Dead",
                    }
                    requiredRequests--
                }
                reqCtx.res <- response
            }
        } else if stage == 4 { //Confirmacion entrada piso 3
            log.Printf("Se han confirmado todos los participantes para continuar al siguiente piso")
            fmt.Printf("Si desea comenzar el piso 3 ingrese cualquier input: ")
            _, _ = reader.ReadString('\n') // Ignora la entrada del usuario
            for i := 0; i < initialRequiredRequests; i++ {
                reqCtx := <-s.reqCh
                response := &pb.MercResponseMessage{
                    Informacion: "3",
                }
                reqCtx.res <- response
            }
            log.Printf("El tercer piso comienza!, se le ha notificado a todos los mercenarios, una vez recibidas sus decisiones se resolvera el tercer piso")
        } else if stage == 5 { //Piso 3
            if Piso3Status[0] < 4 {
                if Piso3Status[1] == -1 { //primera ronda, debe generarse la decision del patriarca
                    Piso3Status[1] = 1 + rand.Intn(15)
                    log.Printf("EL NUMERO DEL PATRIARCA HA SIDO ESTABLECIDO!: %d", Piso3Status[1])
                }
                for i := 0; i < initialRequiredRequests; i++ {
                    reqCtx := <-s.reqCh
                    num, _ := strconv.Atoi(reqCtx.req.Accion)
                    if Piso3Status[1] == num { // Acierto!
                        Piso3Status[2]++
                    }
                    response := &pb.MercResponseMessage{
                        Informacion: "3",
                    }
                    reqCtx.res <- response
                }
                Piso3Status[0]++
            } else {
                if Piso3Status[2]>= 2 {
                    log.Printf("Los Mercenarios han completado el piso 3 con un total de aciertos de %d", Piso3Status[2])
                    for i := 0; i < initialRequiredRequests; i++ {
                        reqCtx := <-s.reqCh
                        response := &pb.MercResponseMessage{
                            Informacion: "Win",
                        }
                        reqCtx.res <- response
                    }
                } else {
                    log.Printf("Los Mercenarios se han muerto en el piso 3 con un total de aciertos de %d", Piso3Status[2])
                    for i := 0; i < initialRequiredRequests; i++ {
                        reqCtx := <-s.reqCh
                        response := &pb.MercResponseMessage{
                            Informacion: "Dead",
                        }
                        reqCtx.res <- response
                    }
                    
                }
                log.Printf("Mision Finalizada, reiniciando parametros de servidor!")
                requiredRequests = 8
                Stage = 0
                Piso3Status = [3]int{0, -1, 0,}
            }
        }

    }()
}

func main() {
    // IP y puerto específicos
    address := "localhost:50052"

    lis, err := net.Listen("tcp", address)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    s := &Server{
        reqCh: make(chan requestContext, requiredRequests),
    }

    grpcServer := grpc.NewServer()
    pb.RegisterMercServiceServer(grpcServer, s)
    reflection.Register(grpcServer)

    fmt.Printf("Server is running on %s\n\n", address)
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}