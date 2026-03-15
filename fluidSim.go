package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"time"
)

// Particle represents a single SPH simulation particle (fluid or wall).
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

// parseScene reads an ASCII scene description from r and populates the
// particles slice. It returns the number of particles placed.
// Each non-space, non-newline character becomes two vertically stacked
// particles; '#' characters become immovable wall particles.
func parseScene(r io.Reader, particles []Particle) int {
	xScan := 0
	yScan := 0
	total := 0
	reader := bufio.NewReader(r)
	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			break
		}
		switch char {
		case '\n':
			yScan += 2
			xScan = -1
		case ' ':
			// no particle at this position
		case '#':
			particles[total].Wallflag = 1
			particles[total+1].Wallflag = 1
			fallthrough
		default:
			particles[total].XPos = float64(xScan)
			particles[total].YPos = float64(yScan)
			particles[total+1].XPos = float64(xScan)
			particles[total+1].YPos = float64(yScan + 1)
			total += 2
		}
		xScan++
	}
	return total
}

// stepDensity computes the density for each particle based on its neighbours.
func stepDensity(particles []Particle, n int) {
	for i := 0; i < n; i++ {
		particles[i].Density = float64(particles[i].Wallflag * 9)
		for j := 0; j < n; j++ {
			dx := particles[i].XPos - particles[j].XPos
			dy := particles[i].YPos - particles[j].YPos
			// Interaction radius is 2.0; skip if dist² > 4.
			if distSq := dx*dx + dy*dy; distSq <= 4.0 {
				q := math.Sqrt(distSq)/2.0 - 1.0
				particles[i].Density += q * q
			}
		}
	}
}

// stepForces computes the pressure and viscosity forces for each particle.
func stepForces(particles []Particle, n int, grav, pres, visc float64) {
	for i := 0; i < n; i++ {
		particles[i].YForce = grav
		particles[i].XForce = 0
		for j := 0; j < n; j++ {
			dx := particles[i].XPos - particles[j].XPos
			dy := particles[i].YPos - particles[j].YPos
			if distSq := dx*dx + dy*dy; distSq <= 4.0 {
				q := math.Sqrt(distSq)/2.0 - 1.0
				pressureTerm := q * (3 - particles[i].Density - particles[j].Density) * pres
				particles[i].XForce += (dx*pressureTerm + (particles[i].XVelocity-particles[j].XVelocity)*visc) / particles[i].Density
				particles[i].YForce += (dy*pressureTerm + (particles[i].YVelocity-particles[j].YVelocity)*visc) / particles[i].Density
			}
		}
	}
}

// integrateAndRasterize updates particle positions (fluid only) and writes
// presence bits into screenBuf for the current frame.
// screenBuf must have length width*height and be zeroed before each call.
func integrateAndRasterize(particles []Particle, n int, screenBuf []byte, width, height int) {
	for i := 0; i < n; i++ {
		if particles[i].Wallflag == 0 {
			// Clamp large forces to improve numerical stability.
			divisor := 10.0
			if particles[i].XForce*particles[i].XForce+particles[i].YForce*particles[i].YForce >= 4.2*4.2 {
				divisor = 11.0
			}
			particles[i].XVelocity += particles[i].XForce / divisor
			particles[i].YVelocity += particles[i].YForce / divisor
			particles[i].XPos += particles[i].XVelocity
			particles[i].YPos += particles[i].YVelocity
		}
		x := int(particles[i].XPos)
		y := int(particles[i].YPos / 2)
		idx := x + width*y
		if y >= 0 && y < height-1 && x >= 0 && x < width-1 {
			screenBuf[idx] |= 8
			screenBuf[idx+1] |= 4
			screenBuf[idx+width] |= 2
			screenBuf[idx+width+1] |= 1
		}
	}
}

// buildFrame converts the bit-mask screenBuf into a printable ASCII frame.
// The returned slice has length width*height+1 (includes newlines).
func buildFrame(screenBuf []byte, width, height int) []byte {
	const charset = " '`-.|//,\\|\\_\\/#"
	frame := make([]byte, width*height+1)
	for k := 0; k < width*height; k++ {
		if k%width == width-1 {
			frame[k] = '\n'
		} else {
			frame[k] = charset[screenBuf[k]]
		}
	}
	return frame
}

func main() {
	var (
		gravity       int
		pressure      int
		viscosity     int
		consoleWidth  int
		consoleHeight int
		targetFPS     int
	)

	flag.IntVar(&gravity, "g", 1, "gravity")
	flag.IntVar(&pressure, "p", 4, "pressure")
	flag.IntVar(&viscosity, "v", 7, "viscosity")
	flag.IntVar(&consoleWidth, "w", 80, "console width in characters")
	flag.IntVar(&consoleHeight, "h", 24, "console height in characters")
	flag.IntVar(&targetFPS, "fps", 12, "maximum frames per second")

	flag.Parse()

	maxParticles := consoleWidth * consoleHeight * 2
	particles := make([]Particle, maxParticles)
	screenBuf := make([]byte, consoleWidth*consoleHeight)

	fmt.Print("\x1b[2J")

	totalOfParticles := parseScene(os.Stdin, particles)

	// Pre-convert simulation constants to float64 once, outside the hot loop.
	grav := float64(gravity)
	pres := float64(pressure)
	visc := float64(viscosity)

	// Each frame is a freshly allocated slice sent over the channel so that
	// the rendering goroutine and the simulation goroutine never share memory.
	buffer := make(chan []byte, 200)

	go func() {
		for {
			stepDensity(particles, totalOfParticles)
			stepForces(particles, totalOfParticles, grav, pres, visc)

			for k := range screenBuf {
				screenBuf[k] = 0
			}

			integrateAndRasterize(particles, totalOfParticles, screenBuf, consoleWidth, consoleHeight)
			buffer <- buildFrame(screenBuf, consoleWidth, consoleHeight)
		}
	}()

	// Wait for the simulation goroutine to pre-fill the buffer before
	// starting to display, so the first frames are available immediately.
	time.Sleep(5 * time.Second)

	frameDuration := time.Second / time.Duration(targetFPS)
	for {
		fmt.Print("\x1b[1;1H")
		time.Sleep(frameDuration)
		fmt.Printf("%s", <-buffer)
	}
}
