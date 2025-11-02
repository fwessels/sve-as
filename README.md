# Simple assembler for ARM SVE

`sve-as` provides a Go-based assembler for ARM’s SVE (Scalable Vector Extensions) instruction set, parsing textual mnemonics and operands into opcodes. It understands scalar, vector, predicated, and prefixed forms (including aliases such as mov/lsl/lsr), making it useful for tooling that needs to generate machine code for SVE without relying on external assemblers.

## Example

```
$ more example_arm64.s 
TEXT ·sve_example(SB), $0
    WORD $0x00000000 // add z1.s, p1/m, z1.s, z2.s
    RET
```
```
$ ./sve-as example_arm64.s 
Processing example_arm64.s
```
```
$ more example_arm64.s    
TEXT ·sve_example(SB), $0
    WORD $0x04800441 // add z1.s, p1/m, z1.s, z2.s
    RET
```

## Auto prefixing

`sve-as` will automatically prefix an instruction with a `movprfx` instruction, if so required:

```
$ more example_arm64.s 
TEXT ·sve_example(SB), $0
    WORD $0x00000000 // add z1.s, p1/m, z1.s, z2.s
    WORD $0x00000000 // add z3.s, p1/m, z1.s, z2.s
    WORD $0x00000000 // add z1.s, p1/z, z1.s, z2.s
    RET
```
```
$ ./sve-as example_arm64.s 
Processing example_arm64.s
```
```
TEXT ·sve_example(SB), $0
    WORD $0x04800441          // add z1.s, p1/m, z1.s, z2.s
    DWORD $0x0480044304912423 // add z3.s, p1/m, z1.s, z2.s
    DWORD $0x0480044104902421 // add z1.s, p1/z, z1.s, z2.s
    RET
```