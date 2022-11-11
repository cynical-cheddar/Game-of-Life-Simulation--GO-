package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	)

func worker(height int,p golParams, c chan byte, aboveChan chan byte, belowChan chan byte,commandChan chan int) {
    world := make([][]byte,height)
    for i :=range world {
        world[i] = make([]byte,p.imageWidth)
    }
    newWorld := make([][]byte,height-2)
        for i :=range newWorld {
            newWorld[i] = make([]byte,p.imageWidth)
        }
    for y:=1;y<height-1;y++ {
                for x:=0;x<p.imageWidth;x++ {
                    world[y][x]=<-c
                }
            }
    running := true
    //paused := false
    for turns := 0; running&&(turns < p.turns); turns++ {

        for x:=0;x<p.imageWidth;x++ {
            belowChan<-world[height-2][x]
        }
        for x:=0;x<p.imageWidth;x++ {
            world[0][x]=<-aboveChan
        }
        for x:=0;x<p.imageWidth;x++ {
            aboveChan<-world[1][x]
        }
        for x:=0;x<p.imageWidth;x++ {
            world[height-1][x]=<-belowChan
        }

        for y := 1; y < height-1; y++ {
      		for x := 0; x < p.imageWidth; x++ {
      			// Placeholder for the actual Game of Life logic: flips alive cells to dead and dead cells to alive.
       			 var neighbors=0
       			 for i:=-1;i<2;i++ {
       			    for j:=-1;j<2;j++ {
                           if(i!=0 || j!=0) {
                               var yb=y+i
                               var xb=(x+j+p.imageWidth)%p.imageWidth
                               if(world[yb][xb]!=0){
                                   neighbors+=1
                               }
                           }
                       }
       			 }
                    if(neighbors==3){
                        newWorld[y-1][x] = 1
                    } else if(neighbors == 2) {
                        newWorld[y-1][x] = world[y][x]
                    } else {
                        newWorld[y-1][x] = 0
                    }
       	   	}
       	}

       	for y:=1;y<height-1;y++ {
            for x:=0;x<p.imageWidth;x++ {
                world[y][x] = newWorld[y-1][x]
            }
        }

        select{
        case command := <-commandChan:
        switch(command) {
            case 0:
            case 1:
                    fmt.Println("COMMAND 1")

            totalAlive := 0
            for y:=1;y<height-1;y++ {
               for x:=0;x<p.imageWidth;x++ {
                  if(world[y][x] != 0) {
                      totalAlive+=1
                  }
               }
            }
            commandChan<-totalAlive
            }
        default:
        }

    }
    fmt.Println("GOT HERE SENDING 9")
    commandChan<-9
                            fmt.Println("GOT HERE A")

    for y:=0;y<height-2;y++ {
        for x:=0;x<p.imageWidth;x++ {
            c<-newWorld[y][x]
        }
    }
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p golParams, d distributorChans, alive chan []cell) {

	// Create the 2D slice to store the world.
	world := make([][]byte, p.imageHeight)
	for i := range world {
		world[i] = make([]byte, p.imageWidth)
	}

    newWorld := make([][]byte, p.imageHeight)
    for i := range world {
    	newWorld[i] = make([]byte, p.imageWidth)
    }
	// Request the io goroutine to read in the image with the given filename.
	d.io.command <- ioInput
	d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight)}, "x")

    workerChannels := make([]chan byte,p.threads)
    haloChannels := make([]chan byte,p.threads)
    workerCommands := make([]chan int,p.threads)    //used in the distributor to tell the workers to do certain workerCommands
    //0-pause
    //1-request totalAlive

    for i := range workerChannels {
       workerChannels[i] = make(chan byte)
       workerCommands[i] = make(chan int,1)
       haloChannels[i] = make(chan byte,p.imageWidth)
    }
    defaultHeight := p.imageHeight/p.threads
    height := make([]int, p.threads)

    for i,c:=range workerChannels {
        if(i!=p.threads-1) {
            height[i] = defaultHeight
        } else {
            height[i]=p.imageHeight-(defaultHeight*(p.threads-1))
        }
        go worker(height[i]+2,p,c,haloChannels[(i-1+p.threads)%p.threads],haloChannels[i],workerCommands[i])
    }


	// The io goroutine sends the requested image byte by byte, in rows.
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			val := <-d.io.inputVal
			if val != 0 {
				fmt.Println("Alive cell at", x, y)
				world[y][x] = val
			}
		}
	}
    for i,c:=range workerChannels {
        if(i != p.threads-1) {
            for y:=((i)*(defaultHeight));y<((i+1)*(defaultHeight));y++ {
                for x:=0;x<p.imageWidth;x++ {
                    c<-world[(y+p.imageHeight)%p.imageHeight][x]
                }
            }
        } else {
            for y:=((i)*(defaultHeight));y<(p.imageHeight);y++ {
                for x:=0;x<p.imageWidth;x++ {
                    c<-world[(y+p.imageHeight)%p.imageHeight][x]
                }
            }
        }
    }
	periodicChan := make(chan bool)
	go ticker(periodicChan)

    running := true
    paused := false
    aliveCount := 0
    finishedCount := 0
	// Calculate the new state of Game of Life after the given number of turns.
	for (running) {
             for y := 0; y < p.imageHeight; y++ {
                for x := 0; x < p.imageWidth; x++ {
                   world[y][x] = newWorld[y][x]
                   if(world[y][x] != 0) {
                       aliveCount+=1
                   }
                }
             }
             select {
                case <-periodicChan:
                       fmt.Println("TRYING TO SEND")

                aliveCount = 0
                //for _,c:=range workerCommands  {
                     //c<-1
                 //            fmt.Println("SENT COMMAND")

                     //aliveCount+=<-c
                //}
                fmt.Println(aliveCount)
                default:
             }
             select{
                 case currentKey := <-d.keyChan:
                     switch(currentKey) {
                         case 's':
                             printBoard(p,d,world)
                         case 'q':
                           //  printBoard(p,d,world)
                             running=false

                         case 'p':
                            paused=true
                            fmt.Println("PAUSED")
                            for (paused) {
                                currentKey2:=<-d.keyChan
                                switch(currentKey2){
                                    case 'p':paused = false
                                }
                            }
                            fmt.Println("UNPAUSED")
                     }
                 default:
        }
        for _,c:=range workerCommands {
            select{
                case command := <-c:
                switch (command) {
                    case 9:
                    finishedCount+=1

                }
                default:
            }

        }
        fmt.Println(finishedCount)
        if(finishedCount == p.threads){
            running =false
        }
	}
	                            fmt.Println("GOT HEREB")

    for i,c:=range workerChannels {
                 if(i != p.threads-1) {
                    for y:=((i)*(defaultHeight));y<((i+1)*(defaultHeight));y++ {
                        for x:=0;x<p.imageWidth;x++ {
                            newWorld[y][x]=<-c
                        }
                    }
                 } else {
                    for y:=((i)*(defaultHeight));y<(p.imageHeight);y++ {
                        for x:=0;x<p.imageWidth;x++ {
                            newWorld[y][x]=<-c
                        }
                    }
                 }
             }
        d.io.command <- ioOutput
        d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight)}, "x")

        // Create an empty slice to store coordinates of cells that are still alive after p.turns are done.
        var finalAlive []cell
        // Go through the world and append the cells that are still alive.
        for y := 0; y < p.imageHeight; y++ {
            for x := 0; x < p.imageWidth; x++ {
                d.io.outputVal<-world[y][x]
                if world[y][x] != 0 {

                    finalAlive = append(finalAlive, cell{x: x, y: y})
                }

            }
        }

        // Make sure that the Io has finished any output before exiting.
        d.io.command <- ioCheckIdle
        <-d.io.idle

        // Return the coordinates of cells that are still alive.
        alive <- finalAlive

}

func ticker(aliveChan chan bool) {
    for{
        time.Sleep(2*time.Second)
        aliveChan<-true
    }
}

func printBoard(p golParams, d distributorChans,world[][]byte) {
    d.io.command <- ioOutput
    d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight)}, "x")
    for y := 0; y < p.imageHeight; y++ {
        for x := 0; x < p.imageWidth; x++ {
       	    d.io.outputVal<-world[y][x]
       	}
    }
    d.io.command <- ioCheckIdle
    <-d.io.idle
}
