package versiondb

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"
)

const data = `
a2380aff030a20ff2dd95def59a0c5265f30768a1c6fbdab519c96fb98df
33d524a94575bac1e91292030a02080b120c63726f6e6f735f3737372d31
1802220c08d9a39a9a0610b8d89eca022a480a202d8f46f61152696c6dc8
0c6367c7ddc926dc3ed7ab23bd37fb7ca83eb00ba207122408011220df77
0b6cb0a9694f550ae515d5fb0ebc08902e0fa47ceec816928e74fd8fee44
322004943bc3709104b52f30f83ca3d30d8bf57551edd48797aef5cee50b
b532c15c3a20e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934c
a495991b7852b8554220f537a6e0561fa0edd12b30ec9b6479e659f6f1fa
1587e69556201bfaf4cf97404a20f537a6e0561fa0edd12b30ec9b6479e6
59f6f1fa1587e69556201bfaf4cf97405220252fe7cf36dd1bb85dafc47a
08961df0cfd8c027defa5e01e958be121599db9d5a209048462db52bb809
6c92f8733fcf26455a000a550d9f3748dddadcc2267917146220e3b0c442
98fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b8556a20
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852
b8557214ae7ccd8d599769209074b81d0c2f3f28624742b71a4612210a1d
0a142861b93c776d100695688250c5b1ba4d44ef2a791880a094a58d1d10
0112210a1d0a14ae7ccd8d599769209074b81d0c2f3f28624742b71880a0
94a58d1d100112ae100a630a0d636f696e5f726563656976656412360a08
7265636569766572122a637263316d33683330776c767366386c6c727578
7470756b64767379306b6d326b756d386c3061773437121a0a06616d6f75
6e74121034313139343530383537377374616b650a5c0a08636f696e6261
736512340a066d696e746572122a637263316d33683330776c767366386c
6c7275787470756b64767379306b6d326b756d386c3061773437121a0a06
616d6f756e74121034313139343530383537377374616b650a5f0a0a636f
696e5f7370656e7412350a077370656e646572122a637263316d33683330
776c767366386c6c7275787470756b64767379306b6d326b756d386c3061
773437121a0a06616d6f756e74121034313139343530383537377374616b
650a630a0d636f696e5f726563656976656412360a087265636569766572
122a637263313778706676616b6d32616d67393632796c73366638347a33
6b656c6c3863356c3838656b6572121a0a06616d6f756e74121034313139
343530383537377374616b650a95010a087472616e7366657212370a0972
6563697069656e74122a637263313778706676616b6d32616d6739363279
6c73366638347a336b656c6c3863356c3838656b657212340a0673656e64
6572122a637263316d33683330776c767366386c6c7275787470756b6476
7379306b6d326b756d386c3061773437121a0a06616d6f756e7412103431
3139343530383537377374616b650a3f0a076d65737361676512340a0673
656e646572122a637263316d33683330776c767366386c6c727578747075
6b64767379306b6d326b756d386c30617734370aa2010a046d696e741224
0a0c626f6e6465645f726174696f1214302e393939393939393739343032
37343439353212210a09696e666c6174696f6e1214302e31323939393939
3739373130313635333031123a0a11616e6e75616c5f70726f766973696f
6e7312253235393939393936343737353631363138382e37363031383234
363033383333383838343312150a06616d6f756e74120b34313139343530
383537370a5f0a0a636f696e5f7370656e7412350a077370656e64657212
2a637263313778706676616b6d32616d67393632796c73366638347a336b
656c6c3863356c3838656b6572121a0a06616d6f756e7412103832333839
3031393532307374616b650a630a0d636f696e5f72656365697665641236
0a087265636569766572122a637263316a7636357333677271663676366a
6c33647034743663397439726b3939636438737037326d70121a0a06616d
6f756e74121038323338393031393532307374616b650a95010a08747261
6e7366657212370a09726563697069656e74122a637263316a7636357333
677271663676366a6c33647034743663397439726b393963643873703732
6d7012340a0673656e646572122a637263313778706676616b6d32616d67
393632796c73366638347a336b656c6c3863356c3838656b6572121a0a06
616d6f756e74121038323338393031393532307374616b650a3f0a076d65
737361676512340a0673656e646572122a637263313778706676616b6d32
616d67393632796c73366638347a336b656c6c3863356c3838656b65720a
7f0a0f70726f706f7365725f726577617264122c0a06616d6f756e741222
343131393435303937362e30303030303030303030303030303030303073
74616b65123e0a0976616c696461746f72123163726376616c6f70657231
326c756b75367578656868616b303270793472637a36357a753073776837
776a36756c726c670a790a0a636f6d6d697373696f6e122b0a06616d6f75
6e7412213431313934353039372e36303030303030303030303030303030
30307374616b65123e0a0976616c696461746f72123163726376616c6f70
657231326c756b75367578656868616b303270793472637a36357a753073
776837776a36756c726c670a770a0772657761726473122c0a06616d6f75
6e741222343131393435303937362e303030303030303030303030303030
3030307374616b65123e0a0976616c696461746f72123163726376616c6f
70657231326c756b75367578656868616b303270793472637a36357a7530
73776837776a36756c726c670a7a0a0a636f6d6d697373696f6e122c0a06
616d6f756e741222333833313038393430372e3638303030303030303030
303030303030307374616b65123e0a0976616c696461746f721231637263
76616c6f70657231326c756b75367578656868616b303270793472637a36
357a753073776837776a36756c726c670a780a0772657761726473122d0a
06616d6f756e74122333383331303839343037362e383030303030303030
3030303030303030307374616b65123e0a0976616c696461746f72123163
726376616c6f70657231326c756b75367578656868616b30327079347263
7a36357a753073776837776a36756c726c670a7a0a0a636f6d6d69737369
6f6e122c0a06616d6f756e741222333833313038393430372e3638303030
303030303030303030303030307374616b65123e0a0976616c696461746f
72123163726376616c6f70657231387a367133386d68767473767972356d
616b38666a38733867346777376b6a6a7030653064680a780a0772657761
726473122d0a06616d6f756e74122333383331303839343037362e383030
3030303030303030303030303030307374616b65123e0a0976616c696461
746f72123163726376616c6f70657231387a367133386d68767473767972
356d616b38666a38733867346777376b6a6a7030653064680a250a0a6665
655f6d61726b657412170a08626173655f666565120b3736353632353030
3030301aa9100aa4040aa1040afa020aba020a2a2f636f736d6f732e7374
616b696e672e763162657461312e4d736743726561746556616c69646174
6f72128b020a070a056e6f646530123b0a12313030303030303030303030
30303030303012123230303030303030303030303030303030301a113130
3030303030303030303030303030301a0131222a63726331326c756b7536
7578656868616b303270793472637a36357a753073776837776a73727730
70702a3163726376616c6f70657231326c756b75367578656868616b3032
70793472637a36357a753073776837776a36756c726c6732430a1d2f636f
736d6f732e63727970746f2e656432353531392e5075624b657912220a20
9b86839433f4229b7b7b51554a25ed32549126eb80fe53497ae336a1e67b
66843a1c0a057374616b6512133130303030303030303030303030303030
3030123b3439643130313933363334303730393664303337396438643863
3833336438626633396334383866403139322e3136382e302e35343a3236
363536125f0a570a4f0a282f65746865726d696e742e63727970746f2e76
312e657468736563703235366b312e5075624b657912230a21026e710a62
a342de0ed4d7c4532dcbcbbafbf19652ed67b237efab70e8b207efac1204
0a020801120410c09a0c1a4119ffd1b342c44e183b8113561595186e1f72
c7273ac4e9ccb04c2b8a4d31b3780eff63cb490102ed39f7482084b4ddb8
4d672ee2607262a312724f0e5b2b79aa0012ff0b123612340a322f636f73
6d6f732e7374616b696e672e763162657461312e4d736743726561746556
616c696461746f72526573706f6e73651ae0055b7b226d73675f696e6465
78223a302c226576656e7473223a5b7b2274797065223a22636f696e5f72
65636569766564222c2261747472696275746573223a5b7b226b6579223a
227265636569766572222c2276616c7565223a22637263317479676d7333
78686873337976343837706878336477346139356a6e3774376c6b393067
6161227d2c7b226b6579223a22616d6f756e74222c2276616c7565223a22
313030303030303030303030303030303030307374616b65227d5d7d2c7b
2274797065223a22636f696e5f7370656e74222c22617474726962757465
73223a5b7b226b6579223a227370656e646572222c2276616c7565223a22
63726331326c756b75367578656868616b303270793472637a36357a7530
73776837776a737277307070227d2c7b226b6579223a22616d6f756e7422
2c2276616c7565223a223130303030303030303030303030303030303073
74616b65227d5d7d2c7b2274797065223a226372656174655f76616c6964
61746f72222c2261747472696275746573223a5b7b226b6579223a227661
6c696461746f72222c2276616c7565223a2263726376616c6f7065723132
6c756b75367578656868616b303270793472637a36357a75307377683777
6a36756c726c67227d2c7b226b6579223a22616d6f756e74222c2276616c
7565223a22313030303030303030303030303030303030307374616b6522
7d5d7d2c7b2274797065223a226d657373616765222c2261747472696275
746573223a5b7b226b6579223a22616374696f6e222c2276616c7565223a
222f636f736d6f732e7374616b696e672e763162657461312e4d73674372
6561746556616c696461746f72227d2c7b226b6579223a226d6f64756c65
222c2276616c7565223a227374616b696e67227d2c7b226b6579223a2273
656e646572222c2276616c7565223a2263726331326c756b753675786568
68616b303270793472637a36357a753073776837776a737277307070227d
5d7d5d7d5d28ffffffffffffffffff0130bca10d3a440a02747812050a03
66656512370a096665655f7061796572122a63726331326c756b75367578
656868616b303270793472637a36357a753073776837776a737277307070
3a3d0a02747812370a076163635f736571122c63726331326c756b753675
78656868616b303270793472637a36357a753073776837776a7372773070
702f303a6b0a02747812650a097369676e6174757265125847662f527330
4c455468673767524e57465a55596268397978796336784f6e4d73457772
696b30787333674f2f32504c5351454337546e3353434345744e32345457
6375346d427959714d53636b384f577974357167413d3a3f0a076d657373
61676512340a06616374696f6e122a2f636f736d6f732e7374616b696e67
2e763162657461312e4d736743726561746556616c696461746f723a670a
0a636f696e5f7370656e7412350a077370656e646572122a63726331326c
756b75367578656868616b303270793472637a36357a753073776837776a
73727730707012220a06616d6f756e741218313030303030303030303030
303030303030307374616b653a6b0a0d636f696e5f726563656976656412
360a087265636569766572122a637263317479676d733378686873337976
343837706878336477346139356a6e3774376c6b393067616112220a0661
6d6f756e741218313030303030303030303030303030303030307374616b
653a760a106372656174655f76616c696461746f72123e0a0976616c6964
61746f72123163726376616c6f70657231326c756b75367578656868616b
303270793472637a36357a753073776837776a36756c726c6712220a0661
6d6f756e741218313030303030303030303030303030303030307374616b
653a520a076d65737361676512110a066d6f64756c6512077374616b696e
6712340a0673656e646572122a63726331326c756b75367578656868616b
303270793472637a36357a753073776837776a7372773070701aa9100aa4
040aa1040afa020aba020a2a2f636f736d6f732e7374616b696e672e7631
62657461312e4d736743726561746556616c696461746f72128b020a070a
056e6f646531123b0a123130303030303030303030303030303030301212
3230303030303030303030303030303030301a1131303030303030303030
303030303030301a0131222a63726331387a367133386d68767473767972
356d616b38666a38733867346777376b6a6a747367726e372a3163726376
616c6f70657231387a367133386d68767473767972356d616b38666a3873
3867346777376b6a6a70306530646832430a1d2f636f736d6f732e637279
70746f2e656432353531392e5075624b657912220a2072b50cf0ed1863ff
c937af99b6ad779a2c223e59459eab7768bda7c2da6f836e3a1c0a057374
616b65121331303030303030303030303030303030303030123b35396566
333139663464383334396466626532613233653131356163353835356430
303938613065403139322e3136382e302e35343a3236363536125f0a570a
4f0a282f65746865726d696e742e63727970746f2e76312e657468736563
703235366b312e5075624b657912230a210242785a75074452d62a6ac222
70ffb8fb01c9375d0ba72887ae800dc619315d1b12040a020801120410c0
9a0c1a41e7019fd760970e02f8967aa0f9820c0b98de32d8e72601aa34fe
60df52356d19591eb2bd8516037d2c52c22170ca533abf72d50d4c7f770d
1d5e045df51ff89c0112ff0b123612340a322f636f736d6f732e7374616b
696e672e763162657461312e4d736743726561746556616c696461746f72
526573706f6e73651ae0055b7b226d73675f696e646578223a302c226576
656e7473223a5b7b2274797065223a22636f696e5f726563656976656422
2c2261747472696275746573223a5b7b226b6579223a2272656365697665
72222c2276616c7565223a22637263317479676d73337868687333797634
3837706878336477346139356a6e3774376c6b3930676161227d2c7b226b
6579223a22616d6f756e74222c2276616c7565223a223130303030303030
30303030303030303030307374616b65227d5d7d2c7b2274797065223a22
636f696e5f7370656e74222c2261747472696275746573223a5b7b226b65
79223a227370656e646572222c2276616c7565223a2263726331387a3671
33386d68767473767972356d616b38666a38733867346777376b6a6a7473
67726e37227d2c7b226b6579223a22616d6f756e74222c2276616c756522
3a22313030303030303030303030303030303030307374616b65227d5d7d
2c7b2274797065223a226372656174655f76616c696461746f72222c2261
747472696275746573223a5b7b226b6579223a2276616c696461746f7222
2c2276616c7565223a2263726376616c6f70657231387a367133386d6876
7473767972356d616b38666a38733867346777376b6a6a70306530646822
7d2c7b226b6579223a22616d6f756e74222c2276616c7565223a22313030
303030303030303030303030303030307374616b65227d5d7d2c7b227479
7065223a226d657373616765222c2261747472696275746573223a5b7b22
6b6579223a22616374696f6e222c2276616c7565223a222f636f736d6f73
2e7374616b696e672e763162657461312e4d736743726561746556616c69
6461746f72227d2c7b226b6579223a226d6f64756c65222c2276616c7565
223a227374616b696e67227d2c7b226b6579223a2273656e646572222c22
76616c7565223a2263726331387a367133386d68767473767972356d616b
38666a38733867346777376b6a6a747367726e37227d5d7d5d7d5d28ffff
ffffffffffffff0130fc8c0d3a440a02747812050a0366656512370a0966
65655f7061796572122a63726331387a367133386d68767473767972356d
616b38666a38733867346777376b6a6a747367726e373a3d0a0274781237
0a076163635f736571122c63726331387a367133386d6876747376797235
6d616b38666a38733867346777376b6a6a747367726e372f303a6b0a0274
7812650a097369676e61747572651258357747663132435844674c346c6e
71672b59494d43356a654d746a6e4a6747714e5035673331493162526c5a
48724b39685259446653785377694677796c4d3676334c564455782f6477
30645867526439522f346e41453d3a3f0a076d65737361676512340a0661
6374696f6e122a2f636f736d6f732e7374616b696e672e76316265746131
2e4d736743726561746556616c696461746f723a670a0a636f696e5f7370
656e7412350a077370656e646572122a63726331387a367133386d687674
73767972356d616b38666a38733867346777376b6a6a747367726e371222
0a06616d6f756e7412183130303030303030303030303030303030303073
74616b653a6b0a0d636f696e5f726563656976656412360a087265636569
766572122a637263317479676d7333786868733379763438377068783364
77346139356a6e3774376c6b393067616112220a06616d6f756e74121831
3030303030303030303030303030303030307374616b653a760a10637265
6174655f76616c696461746f72123e0a0976616c696461746f7212316372
6376616c6f70657231387a367133386d68767473767972356d616b38666a
38733867346777376b6a6a70306530646812220a06616d6f756e74121831
3030303030303030303030303030303030307374616b653a520a076d6573
7361676512110a066d6f64756c6512077374616b696e6712340a0673656e
646572122a63726331387a367133386d68767473767972356d616b38666a
38733867346777376b6a6a747367726e37220208022aec0212260a090880
804010e0aeee26120e08a08d0612040880c60a188080401a090a07656432
353531391a9a020a0b626c6f636b5f626c6f6f6d128a020a05626c6f6f6d
128002000000000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000000000000000
000000000000000000000000000000000000001a250a09626c6f636b5f67
6173120b0a06686569676874120132120b0a06616d6f756e741201303222
1220b995a75e975242574ff1730feaf1e9255743e86ff9d949b04e2906ad
fd10f824230a0462616e6b1a06007374616b652213323030303030303038
32333839303139353230300a0462616e6b1a1b021493354845030274cd4b
f1686abd60ab28ec52e1a77374616b65220b383233383930313935323025
0a0462616e6b10011a1b0214dc6f17bbec824fff8f86587966b2047db6ab
73677374616b65250a0462616e6b10011a1b0214f1829676db577682e944
fc3493d451b67ff3e29f7374616b65270a0462616e6b1a1c037374616b65
001493354845030274cd4bf1686abd60ab28ec52e1a7220100260a046261
6e6b10011a1c037374616b650014dc6f17bbec824fff8f86587966b2047d
b6ab7367260a0462616e6b10011a1c037374616b650014f1829676db5776
82e944fc3493d451b67ff3e29f3a0a0c646973747269627574696f6e1a01
0022270a250a057374616b65121c31363437373830333930343030303030
303030303030303030303030290a0c646973747269627574696f6e1a0101
22160a14ae7ccd8d599769209074b81d0c2f3f28624742b7500a0c646973
747269627574696f6e1a16021438b4089f7762e0c20e9bed8e991e074550
ef5a5222280a260a057374616b65121d3338333130383934303736383030
303030303030303030303030303030500a0c646973747269627574696f6e
1a16021457f96e6b86cdefdb3d412547816a82e3e0ebf9d222280a260a05
7374616b65121d3432343330333435303532383030303030303030303030
303030303030520a0c646973747269627574696f6e1a16061438b4089f77
62e0c20e9bed8e991e074550ef5a52222a0a260a057374616b65121d3334
343739383034363639313230303030303030303030303030303030100252
0a0c646973747269627574696f6e1a16061457f96e6b86cdefdb3d412547
816a82e3e0ebf9d2222a0a260a057374616b65121d333831383733313035
343735323030303030303030303030303030303010024f0a0c6469737472
69627574696f6e1a16071438b4089f7762e0c20e9bed8e991e074550ef5a
5222270a250a057374616b65121c33383331303839343037363830303030
3030303030303030303030304f0a0c646973747269627574696f6e1a1607
1457f96e6b86cdefdb3d412547816a82e3e0ebf9d222270a250a05737461
6b65121c3432343330333435303532383030303030303030303030303030
3030180a096665656d61726b65741a010122080000000000000000450a04
6d696e741a0100223a0a1231323939393939373937313031363533303112
243235393939393936343737353631363138383736303138323436303338
333338383834332a0a06706172616d731a116665656d61726b65742f4261
7365466565220d223736353632353030303030225b0a08736c617368696e
671a1601142861b93c776d100695688250c5b1ba4d44ef2a7922370a3163
726376616c636f6e73313970736d6a307268643567716439746773666776
7476643666347a7737326e656674666d6730180122005b0a08736c617368
696e671a160114ae7ccd8d599769209074b81d0c2f3f28624742b722370a
3163726376616c636f6e7331346537766d7232656a61356a707972356871
77736374656c397033797773346874333468797a18012200cd070a077374
616b696e671a02503222bd070a92030a02080b120c63726f6e6f735f3737
372d311802220c08d9a39a9a0610b8d89eca022a480a202d8f46f6115269
6c6dc80c6367c7ddc926dc3ed7ab23bd37fb7ca83eb00ba2071224080112
20df770b6cb0a9694f550ae515d5fb0ebc08902e0fa47ceec816928e74fd
8fee44322004943bc3709104b52f30f83ca3d30d8bf57551edd48797aef5
cee50bb532c15c3a20e3b0c44298fc1c149afbf4c8996fb92427ae41e464
9b934ca495991b7852b8554220f537a6e0561fa0edd12b30ec9b6479e659
f6f1fa1587e69556201bfaf4cf97404a20f537a6e0561fa0edd12b30ec9b
6479e659f6f1fa1587e69556201bfaf4cf97405220252fe7cf36dd1bb85d
afc47a08961df0cfd8c027defa5e01e958be121599db9d5a209048462db5
2bb8096c92f8733fcf26455a000a550d9f3748dddadcc2267917146220e3
b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b8
556a20e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca49599
1b7852b8557214ae7ccd8d599769209074b81d0c2f3f28624742b7129102
0a3163726376616c6f70657231326c756b75367578656868616b30327079
3472637a36357a753073776837776a36756c726c6712430a1d2f636f736d
6f732e63727970746f2e656432353531392e5075624b657912220a209b86
839433f4229b7b7b51554a25ed32549126eb80fe53497ae336a1e67b6684
20032a133130303030303030303030303030303030303032253130303030
303030303030303030303030303030303030303030303030303030303030
30303a070a056e6f6465304a00524b0a3b0a123130303030303030303030
3030303030303012123230303030303030303030303030303030301a1131
30303030303030303030303030303030120c08d1a39a9a0610a89eb0a602
5a01311291020a3163726376616c6f70657231387a367133386d68767473
767972356d616b38666a38733867346777376b6a6a70306530646812430a
1d2f636f736d6f732e63727970746f2e656432353531392e5075624b6579
12220a2072b50cf0ed1863ffc937af99b6ad779a2c223e59459eab7768bd
a7c2da6f836e20032a133130303030303030303030303030303030303032
253130303030303030303030303030303030303030303030303030303030
30303030303030303a070a056e6f6465314a00524b0a3b0a123130303030
303030303030303030303030301212323030303030303030303030303030
3030301a113130303030303030303030303030303030120c08d1a39a9a06
10a89eb0a6025a0131
`

func TestReadFileStreamer(t *testing.T) {
	buf, err := hex.DecodeString(strings.Replace(data, "\n", "", -1))
	require.NoError(t, err)

	changeSet, err := ReadFileStreamer(bufio.NewReader(bytes.NewReader(buf)))
	require.NoError(t, err)

	require.Equal(t, 21, len(changeSet))
	expItem := types.StoreKVPair{StoreKey: "bank", Delete: false, Key: []uint8{0x0, 0x73, 0x74, 0x61, 0x6b, 0x65}, Value: []uint8{0x32, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x38, 0x32, 0x33, 0x38, 0x39, 0x30, 0x31, 0x39, 0x35, 0x32, 0x30}}
	require.Equal(t, expItem, changeSet[0])
}
