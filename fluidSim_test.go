package main

import (
	"math"
	"strings"
	"testing"
)

// ── parseScene ────────────────────────────────────────────────────────────────

func TestParseScene_FluidParticles(t *testing.T) {
	// A single '.' should produce 2 fluid particles stacked vertically.
	input := ".\n"
	particles := make([]Particle, 10)
	n := parseScene(strings.NewReader(input), particles)

	if n != 2 {
		t.Fatalf("expected 2 particles, got %d", n)
	}
	if particles[0].Wallflag != 0 || particles[1].Wallflag != 0 {
		t.Error("fluid particles must have Wallflag == 0")
	}
	if particles[0].XPos != 0 || particles[1].XPos != 0 {
		t.Errorf("both particles should be at x=0, got %v / %v", particles[0].XPos, particles[1].XPos)
	}
	if particles[0].YPos != 0 {
		t.Errorf("top particle should be at y=0, got %v", particles[0].YPos)
	}
	if particles[1].YPos != 1 {
		t.Errorf("bottom particle should be at y=1, got %v", particles[1].YPos)
	}
}

func TestParseScene_WallParticles(t *testing.T) {
	input := "#\n"
	particles := make([]Particle, 10)
	n := parseScene(strings.NewReader(input), particles)

	if n != 2 {
		t.Fatalf("expected 2 particles, got %d", n)
	}
	if particles[0].Wallflag != 1 || particles[1].Wallflag != 1 {
		t.Error("wall particles must have Wallflag == 1")
	}
}

func TestParseScene_SpaceSkipped(t *testing.T) {
	// A space character must not produce any particles.
	input := " \n"
	particles := make([]Particle, 10)
	n := parseScene(strings.NewReader(input), particles)

	if n != 0 {
		t.Fatalf("expected 0 particles for space-only input, got %d", n)
	}
}

func TestParseScene_XPositions(t *testing.T) {
	// Three consecutive characters on the same line → x = 0, 1, 2.
	input := "...\n"
	particles := make([]Particle, 20)
	n := parseScene(strings.NewReader(input), particles)

	if n != 6 {
		t.Fatalf("expected 6 particles, got %d", n)
	}
	for i, want := range []float64{0, 1, 2} {
		if particles[i*2].XPos != want {
			t.Errorf("particle pair %d: want XPos %v, got %v", i, want, particles[i*2].XPos)
		}
	}
}

func TestParseScene_MultiLine(t *testing.T) {
	// Second line's particles must be at y = 2/3 (two-row stride).
	input := ".\n.\n"
	particles := make([]Particle, 20)
	n := parseScene(strings.NewReader(input), particles)

	if n != 4 {
		t.Fatalf("expected 4 particles, got %d", n)
	}
	// First line: y = 0 and 1.
	if particles[0].YPos != 0 || particles[1].YPos != 1 {
		t.Errorf("line 1 y positions wrong: %v / %v", particles[0].YPos, particles[1].YPos)
	}
	// Second line: y = 2 and 3.
	if particles[2].YPos != 2 || particles[3].YPos != 3 {
		t.Errorf("line 2 y positions wrong: %v / %v", particles[2].YPos, particles[3].YPos)
	}
}

// ── stepDensity ───────────────────────────────────────────────────────────────

func TestStepDensity_WallParticleBaseValue(t *testing.T) {
	// Wall particles start with density 9 (Wallflag * 9).
	particles := []Particle{
		{XPos: 0, YPos: 0, Wallflag: 1},
	}
	stepDensity(particles, len(particles))
	// Self-interaction contributes (0/2 - 1)² = 1 to density.
	expected := 9.0 + 1.0
	if math.Abs(particles[0].Density-expected) > 1e-9 {
		t.Errorf("wall particle density: want %.6f, got %.6f", expected, particles[0].Density)
	}
}

func TestStepDensity_DistantParticlesNoInteraction(t *testing.T) {
	// Two particles separated by more than 2 units must not interact.
	particles := []Particle{
		{XPos: 0, YPos: 0},
		{XPos: 3, YPos: 0}, // distance = 3 > 2
	}
	stepDensity(particles, len(particles))
	// Self-interaction only: q = 0/2 - 1 = -1, q² = 1. Density = 0 + 1 = 1.
	expected := 1.0
	if math.Abs(particles[0].Density-expected) > 1e-9 {
		t.Errorf("density with distant neighbour: want %.6f, got %.6f", expected, particles[0].Density)
	}
	if math.Abs(particles[1].Density-expected) > 1e-9 {
		t.Errorf("density with distant neighbour: want %.6f, got %.6f", expected, particles[1].Density)
	}
}

func TestStepDensity_CloseParticlesInteract(t *testing.T) {
	// Two particles close together: each should have density > 1.
	particles := []Particle{
		{XPos: 0, YPos: 0},
		{XPos: 1, YPos: 0}, // distance = 1 <= 2
	}
	stepDensity(particles, len(particles))
	// For particle 0: self q=-1, neighbour q = 1/2 - 1 = -0.5, contribution 0.25.
	// Density = 0 + 1 + 0.25 = 1.25.
	expected := 1.25
	if math.Abs(particles[0].Density-expected) > 1e-9 {
		t.Errorf("particle 0 density: want %.6f, got %.6f", expected, particles[0].Density)
	}
}

// ── stepForces ────────────────────────────────────────────────────────────────

func TestStepForces_GravityApplied(t *testing.T) {
	// A single isolated particle should have YForce = gravity (no neighbours).
	particles := []Particle{
		{XPos: 0, YPos: 0, Density: 1},
	}
	stepDensity(particles, len(particles))
	stepForces(particles, len(particles), 2.0, 4.0, 7.0)
	// No neighbours within range except itself (self-interaction q = -1 <= 0, no contribution).
	if math.Abs(particles[0].YForce-2.0) > 1e-9 {
		t.Errorf("gravity: want YForce=2.0, got %.6f", particles[0].YForce)
	}
	if math.Abs(particles[0].XForce) > 1e-9 {
		t.Errorf("no horizontal gravity: want XForce=0, got %.6f", particles[0].XForce)
	}
}

// ── integrateAndRasterize ─────────────────────────────────────────────────────

func TestIntegrateAndRasterize_WallDoesNotMove(t *testing.T) {
	// Wall particles must not change position regardless of forces.
	particles := []Particle{
		{XPos: 5, YPos: 6, Wallflag: 1, XForce: 10, YForce: 10},
	}
	buf := make([]byte, 80*24)
	integrateAndRasterize(particles, 1, buf, 80, 24)
	if particles[0].XPos != 5 || particles[0].YPos != 6 {
		t.Errorf("wall particle moved: pos (%.0f, %.0f)", particles[0].XPos, particles[0].YPos)
	}
}

func TestIntegrateAndRasterize_FluidMoves(t *testing.T) {
	// A fluid particle with a positive force must move.
	p := Particle{XPos: 10, YPos: 10, Wallflag: 0, XForce: 1, YForce: 1}
	particles := []Particle{p}
	buf := make([]byte, 80*24)
	integrateAndRasterize(particles, 1, buf, 80, 24)
	if particles[0].XPos == p.XPos && particles[0].YPos == p.YPos {
		t.Error("fluid particle with non-zero force should have moved")
	}
}

func TestIntegrateAndRasterize_BitsSet(t *testing.T) {
	// A particle at (0,0) should set bits in the top-left 2x2 block.
	particles := []Particle{
		{XPos: 0, YPos: 0, Wallflag: 1}, // wall so it doesn't move
	}
	width, height := 80, 24
	buf := make([]byte, width*height)
	integrateAndRasterize(particles, 1, buf, width, height)

	if buf[0]&8 == 0 {
		t.Error("buf[0] should have bit 8 set")
	}
	if buf[1]&4 == 0 {
		t.Error("buf[1] should have bit 4 set")
	}
	if buf[width]&2 == 0 {
		t.Errorf("buf[%d] should have bit 2 set", width)
	}
	if buf[width+1]&1 == 0 {
		t.Errorf("buf[%d] should have bit 1 set", width+1)
	}
}

// ── buildFrame ────────────────────────────────────────────────────────────────

func TestBuildFrame_NewlinesAtEndOfRows(t *testing.T) {
	width, height := 5, 3
	buf := make([]byte, width*height)
	frame := buildFrame(buf, width, height)

	// Every (width-1)-th character (0-indexed: 4, 9, 14) must be '\n'.
	for row := 0; row < height; row++ {
		idx := row*width + (width - 1)
		if frame[idx] != '\n' {
			t.Errorf("row %d: expected newline at index %d, got %q", row, idx, frame[idx])
		}
	}
}

func TestBuildFrame_SpaceForZeroBits(t *testing.T) {
	width, height := 5, 3
	buf := make([]byte, width*height) // all zeros → space character
	frame := buildFrame(buf, width, height)

	// All non-newline positions (up to width*height) should map to space (charset[0] = ' ').
	for k := 0; k < width*height; k++ {
		b := frame[k]
		if k%width == width-1 {
			continue // newline position
		}
		if b != ' ' {
			t.Errorf("index %d: expected space, got %q", k, b)
		}
	}
}

func TestBuildFrame_Length(t *testing.T) {
	width, height := 10, 5
	buf := make([]byte, width*height)
	frame := buildFrame(buf, width, height)
	want := width*height + 1
	if len(frame) != want {
		t.Errorf("frame length: want %d, got %d", want, len(frame))
	}
}
