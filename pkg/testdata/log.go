package testdata

var Log string = "[FBSDK] makechan: chan=9; elemsize=8; dataqsiz=5\n" +
	"[FBSDK] Chansend: chan=9; elemsize=8; dataqsiz=5; qcount=1; gid=6\n" +
	"[FB] chan: obj=0xc000120000; id=841813590018;\n" +
	"[FBSDK] Chansend: chan=9; elemsize=8; dataqsiz=5; qcount=2; gid=6\n" +
	"[FB] chan: obj=0xc000120000; id=841813590019;\n" +
	"[FBSDK] chanrecv: chan=9; elemsize=8; dataqsiz=5; qcount=1; gid=6\n" +
	"[FB] chan: obj=0xc000120000; id=841813590020;\n" +
	"[FBSDK] chanclose: chan=9; gid=6\n" +
	"[FB] chan: obj=0xc000120000; id=841813590021;\n" +
	"[FB] mutex: obj=1; id=841813590024; locked=1; gid=6\n" +
	"[FB] mutex: obj=1; id=841813590025; locked=0; gid=6\n" +
	"[FB] mutex: obj=1; id=841813590022; locked=1; gid=8\n" +
	"[FB] mutex: obj=1; id=841813590023; locked=0; gid=8\n" +
	"[FBSDK] chanrecv: chan=9; elemsize=8; dataqsiz=5; qcount=0; gid=7\n" +
	"[FB] chan: obj=0xc000120000; id=841813590017;\n" +
	"[FB] mutex: obj=2; id=841813590028; locked=1; gid=6\n" +
	"[FB] mutex: obj=2; id=841813590029; locked=0; gid=6\n"
