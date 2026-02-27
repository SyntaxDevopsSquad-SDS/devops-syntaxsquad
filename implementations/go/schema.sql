DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL UNIQUE,
  password TEXT NOT NULL
);

-- Default admin user, password is 'password' (bcrypt hashed)
INSERT INTO users (username, email, password)
    VALUES ('admin', 'admin@whoknows.com', '$2a$10$v/spwONyDHojGbiU6V36BOcKJ/bSt9kO2pl41JJ/CMo0ZcruhWwvq');

CREATE TABLE IF NOT EXISTS pages (
    title TEXT PRIMARY KEY UNIQUE,
    url TEXT NOT NULL UNIQUE,
    language TEXT NOT NULL CHECK(language IN ('en', 'da')) DEFAULT 'en',
    last_updated TIMESTAMP,
    content TEXT NOT NULL
);

INSERT OR IGNORE INTO pages (title, url, language, content) VALUES
('Fortran',      'http://web.archive.org/web/20081220110619/http://en.wikipedia.org:80/wiki/Fortran',      'en', 'Fortran'),
('Algorithm',    'http://web.archive.org/web/20081217070911/http://en.wikipedia.org:80/wiki/Algorithm',    'en', 'Algorithm'),
('MATLAB',       'http://web.archive.org/web/20090110165251/http://en.wikipedia.org:80/wiki/Matlab',       'en', 'MATLAB'),
('JavaScript',   'http://web.archive.org/web/20081218123622/http://en.wikipedia.org:80/wiki/Javascript',   'en', 'JavaScript'),
('Coq',          'http://web.archive.org/web/20090713042410/http://en.wikipedia.org:80/wiki/Coq',          'en', 'Coq'),
('Caml',         'http://web.archive.org/web/20090416213114/http://en.wikipedia.org:80/wiki/Caml',         'en', 'Caml'),
('OCaml',        'http://web.archive.org/web/20081227125521/http://en.wikipedia.org:80/wiki/Ocaml',        'en', 'OCaml'),
('SPSS',         'http://web.archive.org/web/20081218234138/http://en.wikipedia.org:80/wiki/SPSS',         'en', 'SPSS'),
('Prolog',       'http://web.archive.org/web/20081221010318/http://en.wikipedia.org:80/wiki/Prolog',       'en', 'Prolog'),
('C++',          'http://web.archive.org/web/20081217050014/http://en.wikipedia.org:80/wiki/C++',          'en', 'C++'),
('ActionScript', 'http://web.archive.org/web/20090107131058/http://en.wikipedia.org:80/wiki/Actionscript', 'en', 'ActionScript'),
('AWK',          'http://web.archive.org/web/20090107190613/http://en.wikipedia.org:80/wiki/Awk',          'en', 'AWK'),
('ColdFusion',   'http://web.archive.org/web/20090111235212/http://en.wikipedia.org:80/wiki/Coldfusion',   'en', 'ColdFusion'),
('COBOL',        'http://web.archive.org/web/20090108085047/http://en.wikipedia.org:80/wiki/COBOL',        'en', 'COBOL'),
('Clojure',      'http://web.archive.org/web/20081227074722/http://en.wikipedia.org:80/wiki/Clojure',      'en', 'Clojure'),
('PL/SQL',       'http://web.archive.org/web/20090201182036/http://en.wikipedia.org:80/wiki/PL/SQL',       'en', 'PL/SQL'),
('Verilog',      'http://web.archive.org/web/20090125044559/http://en.wikipedia.org:80/wiki/Verilog',      'en', 'Verilog'),
('Simula',       'http://web.archive.org/web/20090425231506/http://en.wikipedia.org:80/wiki/Simula',       'en', 'Simula'),
('PL/I',         'http://web.archive.org/web/20081204151014/http://en.wikipedia.org./wiki/PL/I',           'en', 'PL/I'),
('S-PLUS',       'http://web.archive.org/web/20090208111104/http://en.wikipedia.org:80/wiki/S-PLUS',       'en', 'S-PLUS'),
('Stata',        'http://web.archive.org/web/20090301110025/http://en.wikipedia.org:80/wiki/Stata',        'en', 'Stata'),
('Smalltalk',    'http://web.archive.org/web/20081221010938/http://en.wikipedia.org:80/wiki/Smalltalk',    'en', 'Smalltalk'),
('ABAP',         'http://web.archive.org/web/20081220211158/http://en.wikipedia.org:80/wiki/ABAP',         'en', 'ABAP'),
('Objective-C',  'http://web.archive.org/web/20081218102909/http://en.wikipedia.org:80/wiki/Objective-C',  'en', 'Objective-C'),
('ChucK',        'http://web.archive.org/web/20081217072511/http://en.wikipedia.org:80/wiki/Chuck',        'en', 'ChucK'),
('Perl',         'http://web.archive.org/web/20081217060939/http://en.wikipedia.org:80/wiki/Perl',         'en', 'Perl'),
('Modula-3',     'http://web.archive.org/web/20080809084011/http://en.wikipedia.org:80/wiki/Modula-3',     'en', 'Modula-3'),
('Modula-2',     'http://web.archive.org/web/20081203191158/http://en.wikipedia.org./wiki/Modula-2',       'en', 'Modula-2'),
('SNOBOL',       'http://web.archive.org/web/20090519190950/http://en.wikipedia.org:80/wiki/Snobol',       'en', 'SNOBOL'),
('PHP',          'http://web.archive.org/web/20090109171033/http://en.wikipedia.org:80/wiki/PHP',          'en', 'PHP'),
('SOAP',         'http://web.archive.org/web/20081218171125/http://en.wikipedia.org:80/wiki/SOAP',         'en', 'SOAP'),
('Seed7',        'http://web.archive.org/web/20060923111105/http://en.wikipedia.org/wiki/Seed7',           'en', 'Seed7'),
('VBA',          'http://web.archive.org/web/20090209103150/http://en.wikipedia.org:80/wiki/VBA',          'en', 'VBA'),
('Robotics',     'http://web.archive.org/web/20081217203613/http://en.wikipedia.org:80/wiki/Robotics',     'en', 'Robotics'),
('Squeak',       'http://web.archive.org/web/20081219132922/http://en.wikipedia.org:80/wiki/Squeak',       'en', 'Squeak'),
('Compiler',     'http://web.archive.org/web/20081219012447/http://en.wikipedia.org:80/wiki/Compiler',     'en', 'Compiler'),
('MapReduce',    'http://web.archive.org/web/20081218210401/http://en.wikipedia.org:80/wiki/MapReduce',    'en', 'MapReduce'),
('LabVIEW',      'http://web.archive.org/web/20081207140237/http://en.wikipedia.org:80/wiki/LabVIEW',      'en', 'LabVIEW'),
('TeX',          'http://web.archive.org/web/20081225050826/http://en.wikipedia.org:80/wiki/TeX',          'en', 'TeX'),
('Tcl',          'http://web.archive.org/web/20090102220513/http://en.wikipedia.org:80/wiki/Tcl',          'en', 'Tcl'),
('PowerShell',   'http://web.archive.org/web/20081110171826/http://en.wikipedia.org:80/wiki/PowerShell',   'en', 'PowerShell'),
('BCPL',         'http://web.archive.org/web/20081207055839/http://en.wikipedia.org:80/wiki/BCPL',         'en', 'BCPL'),
('VHDL',         'http://web.archive.org/web/20081221084051/http://en.wikipedia.org:80/wiki/VHDL',         'en', 'VHDL'),
('CouchDB',      'http://web.archive.org/web/20081219102950/http://en.wikipedia.org:80/wiki/CouchDB',      'en', 'CouchDB'),
('ECMAScript',   'http://web.archive.org/web/20081220035659/http://en.wikipedia.org:80/wiki/ECMAScript',   'en', 'ECMAScript'),
('Cryptography', 'http://web.archive.org/web/20090108061116/http://en.wikipedia.org:80/wiki/Cryptography', 'en', 'Cryptography'),
('ALGOL',        'http://web.archive.org/web/20090108184943/http://en.wikipedia.org:80/wiki/ALGOL',        'en', 'ALGOL'),
('REXX',         'http://web.archive.org/web/20090219195713/http://en.wikipedia.org:80/wiki/REXX',         'en', 'REXX'),
('Database',     'http://web.archive.org/web/20081219060743/http://en.wikipedia.org:80/wiki/Database',     'en', 'Database');