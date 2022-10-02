## Go Wrapper / Bindings for EJDB

I found this neat little embeddable DB and wanted to use it in go.


## Requirements

Make sure you have the dev headers of ejdb.

  - https://ejdb.org/
  - https://github.com/Softmotions/ejdb

cgo will be looking for these header files during compilation:

  - stdlib.h
  - ejdb2/ejdb2.h
  - ejdb2/iowow/iwkv.h
  - ejdb2/iowow/iwlog.h
