// ignore_for_file: unnecessary_brace_in_string_interps

import 'dart:html';
import 'dart:async';
import 'dart:convert';
import 'package:intl/intl.dart';

final pollFrequency = const Duration(seconds: 2);

String currentlogfilename = "";
String downloadFile = "";

Timer timer = Timer(const Duration(seconds: 1), () => print('Timer init'));

bool mainTimedOut = false;

String currTest = '';

List selectionModes = [
  'select none',
  "today",
  "the last seven days",
  'select all',
  'sort by siteid',
  'invert selection'
];

hide(String selector) {
  Element? item = querySelector(selector);
  if (item != null) {
    item.hidden = true;
    item.classes.add('hidden');
  }
}

show(String selector) {
  Element? item = querySelector(selector);
  if (item != null) {
    item.classes.remove('hidden');
    item.hidden = false;
  }
}

void main() {
//  hide("#tmx4");
//  hide("#tmx5");
//  hide("#siteidinput");
  currTest = '';
  hide("#testing");
  hide('#sysmaint');
  hide('#messageblock');
  showdate("#dt");

  ButtonElement tmx5button = querySelector("#test_tmx5") as ButtonElement;
  tmx5button.onClick.listen((e) {
    doTmx5(e);
  });

  ButtonElement tmx5nbutton = querySelector("#test_tmx5n") as ButtonElement;
  tmx5nbutton.onClick.listen((e) {
    doTmx5n(e);
  });

  ButtonElement qfambutton = querySelector("#test_Qfam") as ButtonElement;
  qfambutton.onClick.listen((e) {
    doQfam(e);
  });

  ButtonElement tmx4button = querySelector("#test_tmx4") as ButtonElement;
  tmx4button.onClick.listen((e) {
    doTmx4(e);
  });

  ButtonElement tmx1button = querySelector("#test_tmx1") as ButtonElement;
  tmx1button.onClick.listen((e) {
    doTmx1(e);
  });

  ButtonElement tmx3button = querySelector("#test_tmx3") as ButtonElement;
  tmx3button.onClick.listen((e) {
    doTmx3(e);
  });

  ButtonElement files = querySelector("#manage_files") as ButtonElement;

  files.onClick.listen((e) {
    doFiles(e);
  });

  //testbutton.disabled = true;

  ButtonElement maintenance = querySelector("#maintenance") as ButtonElement;

  maintenance.onClick.listen((e) {
    doMaintenance(e);
  });

//  ButtonElement filesbutton = querySelector("#downloads_button");
  // filesbutton.onClick.listen(showfiles);

  window.onResize.listen((e) {
    Element? z = querySelector('#messageblock');
    if (z != null && z.hidden == false) {
      sizeMessageBlock();
    }
  });

  checkstatus();
}

doFiles(event) {
  hide('#main');
  show('#files');
  ButtonElement returnbutton = querySelector("#return_main2") as ButtonElement;
  returnbutton.onClick.listen((e) {
    returnToMain(e);
  });
  showfiles();
  SelectElement modselect = querySelector('#modselect') as SelectElement;
  modselect.value = selectionModes[0];
  modselect.onChange.listen((e) {
    doFileSelection();
  });

  ButtonElement download = querySelector('#download') as ButtonElement;
  download.onClick.listen((Event e) {
    makeZipfile(e);
  });
  ButtonElement archive = querySelector('#archive') as ButtonElement;
  archive.onClick.listen((Event e) {
    archiveFiles(e);
  });
}

doTmx4(event) {
  currTest = 'tmx4';
  show("#testing");
  show('#serial_entry');
  // Element testing = querySelector("#testing")!;
  // testing.classes.remove('hidden');
  enableTest();
//  show("#siteidinput");

  //InputElement siteid = querySelector("#siteid");
  hide('#main');
  rtMainEnable();
}

doTmx1(event) {
  currTest = 'tmx1';
  show("#testing");
  show('#serial_entry');
  // Element testing = querySelector("#testing")!;
  // testing.classes.remove('hidden');
  enableTest();
//  show("#siteidinput");

  //InputElement siteid = querySelector("#siteid");
  hide('#main');
  rtMainEnable();
}

doTmx3(event) {
  currTest = 'tmx3';
  show("#testing");
  show('#serial_entry');
  show('#tmx3_type');
  // Element testing = querySelector("#testing")!;
  // testing.classes.remove('hidden');
  enableTest();
//  show("#siteidinput");

  //InputElement siteid = querySelector("#siteid");
  hide('#main');
  rtMainEnable();
}

enableTest() {
  ButtonElement testbutton = querySelector('#testbutton') as ButtonElement;
  testbutton.text = "Begin test";
  testbutton.disabled = false;
  show('#messageblock');
  InputElement siteid = querySelector('#siteid') as InputElement;
  siteid.disabled = false;
  siteid.focus();
  sizeMessageBlock();
  testbutton.onClick.listen((e) {
    starttest(e);
  });
}

doTmx5(event) {
  show("#testing");
  hide("#serial_entry");
  // Element s = querySelector()!;
  // s.classes.add('hidden');
  // s.hidden = true;

  currTest = 'tmx5';
  // Element testing = querySelector("#testing")!;
  // testing.classes.remove('hidden');
  enableTest();
  InputElement serial = querySelector("#serial") as InputElement;
  serial.value = '';
  hide("#main");
  rtMainEnable();
}

doTmx5n(event) {
  show("#testing");
  hide("#serial_entry");
  // Element s = querySelector()!;
  // s.classes.add('hidden');
  // s.hidden = true;

  currTest = 'tmx5n';
  // Element testing = querySelector("#testing")!;
  // testing.classes.remove('hidden');
  enableTest();
  InputElement serial = querySelector("#serial") as InputElement;
  serial.value = '';
  hide("#main");
  rtMainEnable();
}
doQfam(event) {
  show("#testing");
  hide("#serial_entry");
  // Element s = querySelector()!;
  // s.classes.add('hidden');
  // s.hidden = true;

  currTest = 'Qfam';
  // Element testing = querySelector("#testing")!;
  // testing.classes.remove('hidden');
  enableTest();
  InputElement serial = querySelector("#serial") as InputElement;
  serial.value = '';
  hide("#main");
  rtMainEnable();
}

rtMainEnable() {
  ButtonElement returnbutton = querySelector("#return_main") as ButtonElement;
  returnbutton.onClick.listen((e) {
    returnToMain(e);
  });
}

doMaintenance(event) {
  hide("#main");
  hide('#messageblock');
  Element maint = querySelector("#sysmaint")!;
  maint.classes.remove('hidden');
  show('#sysmaint');
  showdate('#dt1');
  getVersions();
  getAddresses();
  ButtonElement settime = querySelector("#time_fix") as ButtonElement;
  settime.onClick.listen((e) {
    fixTime(e);
  });

  ButtonElement returnbutton = querySelector("#go_main") as ButtonElement;
  returnbutton.onClick.listen((e) {
    returnToMain(e);
  });

  ButtonElement reinitbutton = querySelector("#reinit") as ButtonElement;
  reinitbutton.onClick.listen((e) {
    HttpRequest.getString("/cgi-bin/initcontrolfile.py").then((var resp) {
      window.alert('Control file reinitialized!');
    });
  });
}

getVersions() {
  PreElement? pretag = querySelector('#versions') as PreElement?;
  HttpRequest.getString("/script_ver").then((var resp) {
    pretag!.text = resp;
  });

  Element? s = querySelector('#swver');
  HttpRequest.getString("/version_tag.txt").then((var resp) {
    s!.text = "Software versions: ( $resp )";
  });
}

getAddresses() {
  SpanElement p = querySelector("#ip") as SpanElement;
  while (p.children.isNotEmpty) {
    p.children.removeLast();
  }

  HttpRequest.getString("/cgi-bin/ip_addr.py").then((var resp) {
    Map j = jsonDecode(resp);
    for (String item in j.keys as Iterable<String>) {
      String? address = j[item];
      if (item[0] == 'w') {
        p.appendText('Wireless: $address ');
      } else {
        p.appendText('Wired: $address ');
      }
      p.children.add(BRElement());
    }
  });
}

fixTime(event) {
  DateTime t = DateTime.now().toUtc();
  List s = [t.year, t.month, t.day, t.hour, t.minute, t.second, t.millisecond];

  HttpRequest request = HttpRequest();
  request.onLoadEnd.listen((e) {
    onTimeFix(request, t);
  });
  request.onTimeout.listen((e) {
    mainTimeout();
  });
  String path = '/cgi-bin/settime.py';

  request.open('POST', path);

  String out = jsonEncode({'dt': s});
  request.send(out);
}

onTimeFix(HttpRequest request, DateTime t) {
  String resp = request.responseText!;
  Map j = jsonDecode(resp);
  if (j['resp'] == true) {
    showdate('#dt1');
    window.alert('System time updated to ${t.toLocal()}!');
  } else {
    window.alert('time update failed!');
  }
}

returnToMain(event) {
  document.window!.location.href = '/';
}

clearFiles() {
  SelectElement downloads = querySelector("#downloads") as SelectElement;
  while (downloads.children.isNotEmpty) {
    downloads.children.removeLast();
  }
}

showdate(loc) {
  String path = '/cgi-bin/getdate.py';
  SpanElement? dt = querySelector(loc) as SpanElement?;
  HttpRequest.getString(path).then((var resp) {
    String dtstring = jsonDecode(resp)['datetime'];
    DateTime value = DateTime.parse(dtstring).toLocal();

//    var formatter = new DateFormat('');
    String txt = DateFormat.yMEd().add_jms().format(value);

//    var m = value.minute.toString();
//    if (m.length < 2){
//      m = "0${m}";
//    }
//    String slug = "${value.year.toString()}-${value.month.toString().padLeft(2,'0')}-${value.day.toString().padLeft(2,'0')}";
    dt!.text = txt;
  });
}

showfiles() {
  String path = 'cgi-bin/fileslist.py';
  SelectElement modselect = querySelector('#modselect') as SelectElement;

  if (modselect.children.isEmpty) {
    for (String mode in selectionModes as Iterable<String>) {
      OptionElement o = OptionElement();
      o.text = mode;
      o.value = mode;
      modselect.children.add(o);
    }
  }

  clearFiles();

//  ButtonElement button = event.target;
//  String t = button.text;

  SelectElement? downloads = querySelector("#downloads") as SelectElement?;
  HttpRequest.getString(path).then((var resp) {
    var data = jsonDecode(resp)['files'];

    for (String filename in data) {
      OptionElement o = OptionElement();
      o.text = filename;
      o.value = filename;
      downloads!.children.add(o);

//        TableRowElement r = downloads.addRow();
//        TableCellElement left = r.addCell();
//        TableCellElement right = r.addCell();
//        right.appendText(filename);
//        FormElement f = new FormElement();
//        f.action = "cgi-bin/getfile.py?filename=${filename}";
//        f.method = "POST";
//        ButtonElement b = new ButtonElement();
//        b.appendText("download");
//        f.children.add(b);
//        left.children.add(f);
    }
  });

//  SelectElement modselect = querySelector('#modselect');
  show('#modselect');
}

archiveFiles(Event event) {
  event.preventDefault();
  event.stopPropagation();
  FormElement? theForm = querySelector('#filesform') as FormElement?;
  String action = '/cgi-bin/archive_files.py';
  SelectElement downloads = querySelector("#downloads") as SelectElement;
  int count = 0;
  for (OptionElement o in downloads.children as Iterable<OptionElement>) {
    if (o.selected == true) {
      count += 1;
    }
  }
  bool resp = window.confirm('Moving $count files to Archive. Continue?');
  if (resp == false) {
    return;
  }

  FormData data = FormData(theForm);
  HttpRequest.request(action, method: theForm!.method, sendData: data)
      .then((HttpRequest req) {
    showfiles();
    SelectElement modselect = querySelector('#modselect') as SelectElement;
    modselect.selectedIndex = 0;
  });
}

makeZipfile(Event event) {
  FormElement theForm = querySelector('#filesform') as FormElement;
  theForm.action = '/cgi-bin/get_zip.py';
  SelectElement downloads = querySelector("#downloads") as SelectElement;
  if (downloads.value!.isEmpty) {
    event.preventDefault();
    event.stopPropagation();
    window.alert('Please select files for download,');
    return;
  }
//  HttpRequest request = new HttpRequest();
//
//  request.onLoadEnd.listen((e) {
////    String resp = request.responseText;
//    //Map data = JSON.decode(resp);
//    //if (data['resp'] == true){
//      //get_zipfile();
//    //}
//
//  });
//  request.onTimeout.listen((e) {
//
//  });
//
//  request.setRequestHeader('Content-type', 'application/json');
//
//  String path = 'cgi-bin/get_zip.py';
//
//  request.open('POST', path);
//  //String address = getSiteId();
//  List theList = [];
//  for (OptionElement o in downloads.children)
//    if (o.selected == true){
//      theList.add(o.value);
//    }
//
//  String out = JSON.encode({'files':theList});
//  request.send(out);
//
//
}

doFileSelection() {
  //List SelectionModes = ['select none', "today's files", "this week's files", 'select all'];
  SelectElement? downloads = querySelector("#downloads") as SelectElement?;
  SelectElement modselect = querySelector('#modselect') as SelectElement;
  if (modselect.value == selectionModes[0]) {
    for (OptionElement o in downloads!.children as Iterable<OptionElement>) {
      //select none
      o.selected = false;
    }
  } else if (modselect.value == selectionModes[3]) {
    //select all
    for (OptionElement o in downloads!.children as Iterable<OptionElement>) {
      o.selected = true;
    }
  } else if (modselect.value == selectionModes[1]) {
    //select today
    sortOptions(downloads!, datesort);
    String today = sformat(DateTime.now());
    for (OptionElement o in downloads.children as Iterable<OptionElement>) {
      Map meta = metaParse(o.value);
      if (meta['date'] == today) {
        o.selected = true;
      } else {
        o.selected = false;
      }
    }
  } else if (modselect.value == selectionModes[2]) {
    //select last seven days
    sortOptions(downloads!, datesort);
    List lastseven = lastSevenDays();
    for (OptionElement o in downloads.children as Iterable<OptionElement>) {
      if (lastseven.contains(metaParse(o.value)['date'])) {
        o.selected = true;
      } else {
        o.selected = false;
      }
    }
  } else if (modselect.value == selectionModes[4]) {
    sortOptions(downloads!, sitesort);
  } else if (modselect.value == selectionModes[5]) {
    for (OptionElement o in downloads!.children as Iterable<OptionElement>) {
      if (o.selected == true) {
        o.selected = false;
      } else {
        o.selected = true;
      }
    }
  }
}

sortOptions(SelectElement element, Comparator compare) {
  List myList = [];
  for (OptionElement item in element.children as Iterable<OptionElement>) {
    myList.add(item.value);
  }
  myList.sort(compare);

  while (element.children.isNotEmpty) {
    element.children.removeLast();
  }
  for (String item in myList as Iterable<String>) {
    OptionElement o = OptionElement();
    o.text = item;
    o.value = item;
    element.children.add(o);
  }
}

lastSevenDays() {
  DateTime today = DateTime.now();
  List theList = [today];
  int count = 1;
  while (theList.length < 7) {
    theList.add(today.subtract(Duration(days: count)));
    count += 1;
  }
  List newList = [];
  for (DateTime d in theList as Iterable<DateTime>) {
    newList.add(sformat(d));
  }
  return newList;
}

int sitesort(dynamic elem1, dynamic elem2) {
  String a = metaParse(elem1)['siteid'];
  String b = metaParse(elem2)['siteid'];
  return a.compareTo(b);
}

int datesort(dynamic elem1, dynamic elem2) {
  String a = metaParse(elem1)['date'];
  String b = metaParse(elem2)['date'];
  return -1 * a.compareTo(b);
}

metaParse(String s) {
  List info = s.split('_');
  Map data = {};
  try {
    data['siteid'] = info[0];
    data['date'] = info[1];
  } on RangeError {
    return {'siteid': '9999', 'date': '20140101'};
  }
  return data;
}

sformat(adate) {
  String month = '${adate.month}'.padLeft(2, '0');
  String day = '${adate.day}'.padLeft(2, '0');
  String stoday = "${adate.year}${month}${day}";
  return stoday;
}

//getSiteId() {
//  InputElement t = querySelector("#siteid");
//  String site = t.value;
//  return site;
//}

validateInputs() {
  InputElement t = querySelector("#siteid") as InputElement;

  if (int.tryParse(t.value!) == null) {
    return false;
  }

  num siteid = t.valueAsNumber!;
  num max = int.parse(t.attributes['max']!);
  num min = int.parse(t.attributes['min']!);
  if (siteid > max || siteid < min) {
    return false;
  }
  if (['tmx4', 'tmx3', 'tmx1'].contains(currTest)) {
    InputElement t = querySelector("#serial") as InputElement;
    if (int.tryParse(t.value!) == null) {
      return false;
    }
    num serial = t.valueAsNumber!;
    num max = int.parse(t.attributes['max']!);
    num min = int.parse(t.attributes['min']!);
    if (serial > max || serial < min) {
      return false;
    }
  }
  return true;
}

getInputs() {
  String serial = '';
  InputElement t = querySelector("#siteid") as InputElement;
  String? siteid = t.value;

// address is included for historical reasons
  Map out = {'siteid': siteid, 'address': siteid, 'c': currTest};

  if (['tmx4', 'tmx3', 'tmx1'].contains(currTest)) {
    InputElement s = querySelector("#serial") as InputElement;
    serial = s.value!.padLeft(8, '0');
    out['serial'] = serial;
  }
  if (currTest == 'tmx3') {
    String x3 = "";
    ElementList radios = querySelectorAll('input[name="tmx3"]');
    for (Element item in radios) {
      item = item as RadioButtonInputElement;
      if (item.checked != null) {
        if (item.checked == true) {
          x3 = item.value!;
          break;
        }
      }
    }
    Element ch = querySelector("#channel_input")!;
    ch = ch as TextInputElement;
    String v = "";
    var chv = [];
    if (ch.value != null) {
      v = ch.value!;
      var s = v.split("-");
      for (var i in s) {
        chv.add(int.parse(i));
      }
    }
    if (chv.isNotEmpty) {
      out['channels'] = chv;
    }
    out['x3'] = x3;
  }

  return out;
}

void starttest(event) {
  event.preventDefault();
  ButtonElement? button = event.target;
  bool validated = validateInputs();

  if (!validated) {
    window.alert('Please correct input and try again.');
    return;
  }

  if (button!.text == 'New test') {
    InputElement t = querySelector("#siteid") as InputElement;
    t.value = "";
    t.disabled = false;
    t.focus();
    if (currTest == 'tmx4') {
      InputElement serial = querySelector('#serial') as InputElement;
      serial.disabled = false;
      serial.value = '';
    }
    button.disabled = false;
    button.text = 'Begin test';
  } else {
    button.text = "Test in progress";
    button.disabled = true;
    sizeMessageBlock();

    hide('#file_available');

    DivElement z = querySelector('#messages') as DivElement;
    while (z.children.isNotEmpty) {
      z.children.removeLast();
    }
    var ok = ParagraphElement();
    ok.appendText("OK. Starting test.");
    z.append(ok);

    HttpRequest request = HttpRequest();
    request.onLoadEnd.listen((e) {
      onResponse(request);
    });
    request.onTimeout.listen((e) {
      mainTimeout();
    });
    String path = '/cgi-bin/dotest.py';

    request.open('POST', path);
    //String address = getSiteId();

    String out = jsonEncode(getInputs());
    starttimer();
    request.send(out);

////  HttpRequest.getString(path).then((data) {
//    HeadingElement top = querySelector("#top");
//    top.scrollIntoView();
//  });

    Timer(Duration(seconds: 5), checkstatus);
  }
}

mainTimeout() {
  mainTimedOut = true;
  window.alert("main script timed out (harmless)");
//  Location loc =window.location;
//  document.window.location = loc;
}

onResponse(HttpRequest request) {}

void checkstatus() {
  String path = 'control.json';
  HttpRequest.getString(path).then((var resp) {
    Map? data = {};
    try {
      data = jsonDecode(resp);
    } on FormatException {
      // got incomplete JSON file
      starttimer();
      return;
    }
    if (data!['status'] == 'done' && data['filename'] != null) {
      doDoneStatus(data);
      return;
    } else if (data['status'] == 'new test') {
      starttimer();
      return;
    } else if (data['status'] == 'tests in progress') {
      showmessages(data['messages']);
      if (data['address'] != null) {
        if (currTest != '') {
          InputElement t = querySelector("#siteid") as InputElement;
//        t.value = "Testing - ${data['address']}";
          t.value = data['address'];
          t.disabled = true;
          if (currTest == 'tmx4') {
            InputElement? ie = querySelector("#serial") as InputElement?;
            if (ie != null) {
              var serial = data['serial'];
              if (serial != null && serial != '') {
                ie.value = serial;
              }
              ie.disabled = true;
            }
          }
          ButtonElement? button =
              querySelector("#testbutton") as ButtonElement?;
          if (button != null) {
            button.text = "Test in progress";
            button.disabled = true;
          }
        }
      }
      starttimer();
      return;
    }
    ButtonElement testbutton = querySelector("#testbutton") as ButtonElement;
    testbutton.text = 'Retry';
    testbutton.disabled = false;
  }).catchError(noControlFile);
}

sizeMessageBlock() {
  DivElement msgs = querySelector('#messageblock') as DivElement;
  Element testing = querySelector('#testing')!;
  Element top = querySelector('#top')!;
  //var clientheight = document.window.

  DivElement messageContent = querySelector('#messages') as DivElement;
  int height = testing.clientHeight + top.clientHeight;

  int yavailable = window.innerHeight!;
  int ht = yavailable - height;
  msgs.style.top = "${height}";
  //window.alert("setting content height to ${ht}");
  //msgs.style.height = "{ht-10}px";
  messageContent.style.height = "${ht - 60}px";
}

starttimer() {
  if (!timer.isActive) {
    timer = Timer(pollFrequency, checkstatus);
  }
}

doDoneStatus(data) {
  String? filename = data['filename'];
  if (filename != null) {
    downloadFile = filename;
  }
  //InputElement addressinput = querySelector('#address');
  //addressinput.disabled = true;
  //addressinput.value = "Completed: ${data['address']}";
  show('#file_available');

  ButtonElement testbutton = querySelector("#testbutton") as ButtonElement;

  testbutton.text = "New test";
  testbutton.disabled = false;
  FormElement form = querySelector('#download_latest') as FormElement;
  form.method = "POST";
  form.action = '/cgi-bin/getfile.py?filename=${filename}';
  DivElement msgs = querySelector('#messages') as DivElement;
  sizeMessageBlock();
  if (msgs.children.isNotEmpty) {
    var z = ParagraphElement();
    z.appendText("done");
    msgs.append(z);
    z.scrollIntoView();
    AudioElement? bell = querySelector("#bell") as AudioElement?;
    if (bell != null) {
      bell.play();
    }
  }
//  HeadingElement top = querySelector("#mainbody");
//  top.scrollIntoView();
}

getFile(event) {
  if (downloadFile.length > 3) {
    String download = '/cgi-bin/getfile.py?filename=${downloadFile}';
    HttpRequest.getString(download);
  }
}

noControlFile(event) {
  //window.alert("Warning: Control file not found. (harmless!)");
  //InputElement t = querySelector("#address");
  //t.disabled = true;
  //t.value = 'Press "New test" to begin.';
  window.alert('${event}');
  ButtonElement button = querySelector("#testbutton") as ButtonElement;
  button.disabled = false;
  starttimer();
}

showmessages(List aList) {
  DivElement liststart = querySelector('#messages') as DivElement;
  while (liststart.children.isNotEmpty) {
    liststart.children.removeLast();
  }
  for (var item in aList) {
    if (item is List) {
      DateTime starttime = DateTime.parse(item[2]).toLocal();
      DateTime currtime = DateTime.now();
      DateTime endtime = starttime.add(Duration(seconds: item[1]));
      int diff = endtime.difference(currtime).inSeconds;
      if (diff > 0) {}
      item = item[0];
    }

    var z = ParagraphElement();
    z.appendText(item);
    liststart.append(z);
    sizeMessageBlock();
    z.scrollIntoView();
  }
}
