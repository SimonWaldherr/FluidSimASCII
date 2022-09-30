# FluidSimASCII

This is a [Golang](http://golang.org/) fluid simulator using the "[Smoothed-particle hydrodynamics (SPH)](http://en.wikipedia.org/wiki/Smoothed-particle_hydrodynamics)" method.

## origin

It is based on the C code from [davidedc/Ascii-fluid-simulation-deobfuscated](https://github.com/davidedc/Ascii-fluid-simulation-deobfuscated), 
which is the best de-obfuscated version of [Yusuke Endoh](https://github.com/mame/)'s  
["Most complex ASCII fluid" obfuscated C code competition 2012 entry](http://www.ioccc.org/2012/endoh1/hint.html).  
The original code of Yusuke Endoh is in the [repo of the "International Obfuscated C Code Contest"](https://github.com/ioccc-src/winner).
The code contained there, respectively the entire repository, is under the [CC-0 license](https://github.com/ioccc-src/winner/blob/master/LICENSE).  

## run

You can run it like this:

```sh
go run fluidSim.go < demo01.txt
```

