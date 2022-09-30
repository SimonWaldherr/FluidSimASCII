package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"time"
)

const CONSOLE_WIDTH = 80
const CONSOLE_HEIGHT = 24

var xSandboxAreaScan int = 0
var ySandboxAreaScan int = 0

type Particle struct {
	XPos      float64
	YPos      float64
	Density   float64
	Wallflag  int
	XForce    float64
	YForce    float64
	XVelocity float64
	YVelocity float64
}

var particles [CONSOLE_WIDTH * CONSOLE_HEIGHT * 2]Particle
var xParticleDistance, yParticleDistance, particlesInteraction, particlesDistance float64
var x, y, screenBufferIndex, totalOfParticles int

var gravity, pressure, viscosity int

var screenBuffer [CONSOLE_WIDTH*CONSOLE_HEIGHT + 1]byte
var char rune
var err error

func main() {
	flag.IntVar(&gravity, "g", 1, "gravity")
	flag.IntVar(&pressure, "p", 4, "pressure")
	flag.IntVar(&viscosity, "v", 7, "viscosity")

	flag.Parse()

	fmt.Println("\x1b[2J")
	var particlesCounter int = 0
	reader := bufio.NewReader(os.Stdin)

	for err == nil {
		char, _, err = reader.ReadRune()

		switch char {
		case '\n':
			ySandboxAreaScan += 2
			xSandboxAreaScan = -1
		case ' ':
		case '#':
			p := &particles[particlesCounter+1].Wallflag
			particles[particlesCounter+1].Wallflag = 1
			particles[particlesCounter].Wallflag = *p
			fallthrough
		default:
			particles[particlesCounter].XPos = float64(xSandboxAreaScan)
			particles[particlesCounter].YPos = float64(ySandboxAreaScan)
			particles[particlesCounter+1].XPos = float64(xSandboxAreaScan)
			particles[particlesCounter+1].YPos = float64(ySandboxAreaScan + 1)
			particlesCounter += 2
			totalOfParticles = particlesCounter
		}
		xSandboxAreaScan += 1
	}

	buffer := make(chan [CONSOLE_WIDTH*CONSOLE_HEIGHT + 1]byte, 200)

	go func() {
		for {
			var particlesCursor, particlesCursor2 int

			for particlesCursor = 0; particlesCursor < totalOfParticles; particlesCursor++ {
				particles[particlesCursor].Density = float64(particles[particlesCursor].Wallflag * 9)
				for particlesCursor2 = 0; particlesCursor2 < totalOfParticles; particlesCursor2++ {
					xParticleDistance = particles[particlesCursor].XPos - particles[particlesCursor2].XPos
					yParticleDistance = particles[particlesCursor].YPos - particles[particlesCursor2].YPos
					particlesDistance = math.Sqrt(math.Pow(xParticleDistance, 2.0) + math.Pow(yParticleDistance, 2.0))
					particlesInteraction = particlesDistance/2.0 - 1.0
					if math.Floor(1.0-particlesInteraction) > 0 {
						particles[particlesCursor].Density += particlesInteraction * particlesInteraction
					}
				}
			}

			for particlesCursor = 0; particlesCursor < totalOfParticles; particlesCursor++ {
				particles[particlesCursor].YForce = float64(gravity)
				particles[particlesCursor].XForce = 0
				for particlesCursor2 = 0; particlesCursor2 < totalOfParticles; particlesCursor2++ {
					xParticleDistance = particles[particlesCursor].XPos - particles[particlesCursor2].XPos
					yParticleDistance = particles[particlesCursor].YPos - particles[particlesCursor2].YPos
					particlesDistance = math.Sqrt(math.Pow(xParticleDistance, 2.0) + math.Pow(yParticleDistance, 2.0))
					particlesInteraction = particlesDistance/2.0 - 1.0
					if math.Floor(1.0-particlesInteraction) > 0 {
						particles[particlesCursor].XForce += particlesInteraction * (xParticleDistance*(3-particles[particlesCursor].Density-particles[particlesCursor2].Density)*float64(pressure) + particles[particlesCursor].XVelocity*float64(viscosity) - particles[particlesCursor2].XVelocity*float64(viscosity)) / particles[particlesCursor].Density
						particles[particlesCursor].YForce += particlesInteraction * (yParticleDistance*(3-particles[particlesCursor].Density-particles[particlesCursor2].Density)*float64(pressure) + particles[particlesCursor].YVelocity*float64(viscosity) - particles[particlesCursor2].YVelocity*float64(viscosity)) / particles[particlesCursor].Density
					}
				}
			}

			for screenBufferIndex = 0; screenBufferIndex < int(CONSOLE_WIDTH*CONSOLE_HEIGHT); screenBufferIndex++ {
				screenBuffer[screenBufferIndex] = 0
			}

			for particlesCursor = 0; particlesCursor < totalOfParticles; particlesCursor++ {
				if particles[particlesCursor].Wallflag == 0 {
					if math.Sqrt(math.Pow(particles[particlesCursor].XForce, 2.0)+math.Pow(particles[particlesCursor].YForce, 2.0)) < 4.2 {
						particles[particlesCursor].XVelocity += particles[particlesCursor].XForce / 10
						particles[particlesCursor].YVelocity += particles[particlesCursor].YForce / 10
					} else {
						particles[particlesCursor].XVelocity += particles[particlesCursor].XForce / 11
						particles[particlesCursor].YVelocity += particles[particlesCursor].YForce / 11
					}
					particles[particlesCursor].XPos += particles[particlesCursor].XVelocity
					particles[particlesCursor].YPos += particles[particlesCursor].YVelocity
				}
				x = int(particles[particlesCursor].XPos)
				y = int(particles[particlesCursor].YPos / 2)
				screenBufferIndex = x + CONSOLE_WIDTH*y
				if y >= 0 && y < int(CONSOLE_HEIGHT-1) && x >= 0 && x < int(CONSOLE_WIDTH-1) {
					screenBuffer[screenBufferIndex] |= 8
					screenBuffer[screenBufferIndex+1] |= 4
					screenBuffer[screenBufferIndex+CONSOLE_WIDTH] |= 2
					screenBuffer[screenBufferIndex+CONSOLE_WIDTH+1] |= 1
				}
			}

			for screenBufferIndex = 0; screenBufferIndex < int(CONSOLE_WIDTH*CONSOLE_HEIGHT); screenBufferIndex++ {
				if screenBufferIndex%CONSOLE_WIDTH == int(CONSOLE_WIDTH-1) {
					screenBuffer[screenBufferIndex] = byte('\n')
				} else {
					screenBuffer[screenBufferIndex] = byte(" '`-.|//,\\|\\_\\/#"[screenBuffer[screenBufferIndex]])
				}
			}

			buffer <- screenBuffer
		}
	}()

	time.Sleep(5 * time.Second)

	for {
		fmt.Println("\x1b[1;1H")
		time.Sleep(80 * time.Millisecond)
		fmt.Printf("%s", <-buffer)
	}
}
