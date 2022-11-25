import 'dart:html';
import 'dart:async';
import 'dart:convert';
import 'package:intl/intl.dart';

const POLLFREQUENCY = const Duration(seconds: 2);

String currentlogfilename;
String downloadFile;

bool mainTimedOut = false;

String CurrTest = '';

List SelectionModes = [
  'select none',
  "today",
  "the last seven days",
  'select all',
  'sort by siteid',
  'invert selection'
];

hide(String selector) {
  Element item = querySelector(selector);

  item.hidden = true;
}

show(String selector) {
  Element item = querySelector(selector);
  item.classes.remove('hidden');
  item.hidden = false;
}

void main() {
//  hide("#tmx4");
//  hide("#tmx5");
//  hide("#siteidinput");
  hide("#testing");
  hide('#sysmaint');
  hide('#messageblock');
  showdate("#dt");
  ButtonElement tmx5button = querySelector("#test_tmx5");
  tmx5button.onClick.listen((e) {
    do_tmx5(e);
  });
  CurrTest = '';
  ButtonElement tmx4button = querySelector("#test_tmx4");
  tmx4button.onClick.listen((e) {
    do_tmx4(e);
  });

  ButtonElement files = querySelector("#manage_files");

  files.onClick.listen((e) {
    do_files(e);
  });

  //testbutton.disabled = true;

  ButtonElement maintenance = querySelector("#maintenance");

  maintenance.onClick.listen((e) {
    do_maintenance(e);
  });

//  ButtonElement filesbutton = querySelector("#downloads_button");
  // filesbutton.onClick.listen(showfiles);

  window.onResize.listen((e) {
    Element z = querySelector('#messageblock');
    if (z != null && z.hidden == false) {
      sizeMessageBlock();
    }
  });

  checkstatus();
}

do_files(event) {
  hide('#main');
  show('#files');
  ButtonElement returnbutton = querySelector("#return_main2");
  returnbutton.onClick.listen((e) {
    return_main(e);
  });
  showfiles();
  SelectElement modselect = querySelector('#modselect');
  modselect.value = SelectionModes[0];
  modselect.onChange.listen((e) {
    doFileSelection();
  });

  ButtonElement download = querySelector('#download');
  download.onClick.listen((Event e) {
    make_zipfile(e);
  });
  ButtonElement archive = querySelector('#archive');
  archive.onClick.listen((Event e) {
    archiveFiles(e);
  });
}

do_tmx4(event) {
  CurrTest = 'tmx4';
  show("#testing");
  Element testing = querySelector("#testing");
  testing.classes.remove('hidden');
  enableTest();
//  show("#siteidinput");
  show('#tmx4');
  //InputElement siteid = querySelector("#siteid");
  hide('#main');
  rt_main_enable();
}

enableTest() {
  ButtonElement testbutton = querySelector('#testbutton');
  testbutton.text = "Begin test";
  testbutton.disabled = false;
  show('#messageblock');
  InputElement siteid = querySelector('#siteid');
  siteid.disabled = false;
  siteid.focus();
  sizeMessageBlock();
  testbutton.onClick.listen((e) {
    starttest(e);
  });
}

do_tmx5(event) {
  show("#testing");
  CurrTest = 'tmx5';
  Element testing = querySelector("#testing");
  testing.classes.remove('hidden');
  enableTest();
//  show("#siteidinput");
//  show('#tmx5');
  hide("#tmx4");
  InputElement serial = querySelector("#serial");
  serial.value = '';
  hide("#main");
  rt_main_enable();
}

rt_main_enable() {
  ButtonElement returnbutton = querySelector("#return_main");
  returnbutton.onClick.listen((e) {
    return_main(e);
  });
}

do_maintenance(event) {
  hide("#main");
  hide('#messageblock');
  Element maint = querySelector("#sysmaint");
  maint.classes.remove('hidden');
  show('#sysmaint');
  showdate('#dt1');
  getVersions();
  getAddresses();
  ButtonElement settime = querySelector("#time_fix");
  settime.onClick.listen((e) {
    fix_time(e);
  });

  ButtonElement returnbutton = querySelector("#go_main");
  returnbutton.onClick.listen((e) {
    return_main(e);
  });

  ButtonElement reinitbutton = querySelector("#reinit");
  reinitbutton.onClick.listen((e) {
    HttpRequest.getString("/cgi-bin/initcontrolfile.py").then((var resp) {
      window.alert('Control file reinitialized!');
    });
  });
}

getVersions() {
  PreElement pretag = querySelector('#versions');
  HttpRequest.getString("/script_ver").then((var resp) {
    pretag.text = resp;
  });

  Element s = querySelector('#swver');
  HttpRequest.getString("/version_tag.txt").then((var resp) {
    s.text = "Software versions: ( " + resp + " )";
  });
}

getAddresses() {
  ParagraphElement p = querySelector("#ip");
  while (p.children.length > 0) {
    p.children.removeLast();
  }

  HttpRequest.getString("/cgi-bin/ip_addr.py").then((var resp) {
    Map j = jsonDecode(resp);
    for (String item in j.keys) {
      String address = j[item];
      if (item[0] == 'w') {
        p.appendText('Wireless: ${address} ');
      } else {
        p.appendText('Wired: ${address} ');
      }
      p.children.add(new BRElement());
    }
  });
}

fix_time(event) {
  DateTime t = new DateTime.now().toUtc();
  List s = [t.year, t.month, t.day, t.hour, t.minute, t.second, t.millisecond];

  HttpRequest request = new HttpRequest();
  request.onLoadEnd.listen((e) {
    onTimeFix(request, t);
  });
  request.onTimeout.listen((e) {
    mainTimeout();
  });
  String path = 'cgi-bin/settime.py';

  request.open('POST', path);

  String out = jsonEncode({'dt': s});
  request.send(out);
}

onTimeFix(HttpRequest request, DateTime t) {
  String resp = request.responseText;
  Map j = jsonDecode(resp);
  if (j['resp'] == true) {
    showdate('#dt1');
    window.alert('System time updated to ${t.toLocal()}!');
  } else {
    window.alert('time update failed!');
  }
}

return_main(event) {
  document.window.location.href = '/';
}

clearFiles() {
  SelectElement downloads = querySelector("#downloads");
  while (downloads.children.length > 0) {
    downloads.children.removeLast();
  }
}

showdate(loc) {
  String path = 'cgi-bin/getdate.py';
  SpanElement dt = querySelector(loc);
  HttpRequest.getString(path).then((var resp) {
    String dtstring = jsonDecode(resp)['datetime'];
    DateTime value = DateTime.parse(dtstring).toLocal();

//    var formatter = new DateFormat('');
    String txt = new DateFormat.yMEd().add_jms().format(value);

//    var m = value.minute.toString();
//    if (m.length < 2){
//      m = "0${m}";
//    }
//    String slug = "${value.year.toString()}-${value.month.toString().padLeft(2,'0')}-${value.day.toString().padLeft(2,'0')}";
    dt.text = txt;
  });
}

showfiles() {
  String path = 'cgi-bin/fileslist.py';
  SelectElement modselect = querySelector('#modselect');

  if (modselect.children.length == 0) {
    for (String mode in SelectionModes) {
      OptionElement o = new OptionElement();
      o.text = mode;
      o.value = mode;
      modselect.children.add(o);
    }
  }

  clearFiles();

//  ButtonElement button = event.target;
//  String t = button.text;

  SelectElement downloads = querySelector("#downloads");
  HttpRequest.getString(path).then((var resp) {
    var data = jsonDecode(resp)['files'];

    for (String filename in data) {
      OptionElement o = new OptionElement();
      o.text = filename;
      o.value = filename;
      downloads.children.add(o);

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
  FormElement theForm = querySelector('#filesform');
  String action = 'cgi-bin/archive_files.py';
  SelectElement downloads = querySelector("#downloads");
  int count = 0;
  for (OptionElement o in downloads.children) {
    if (o.selected == true) {
      count += 1;
    }
  }
  bool resp = window.confirm('Moving ${count} files to Archive. Continue?');
  if (resp == false) {
    return;
  }

  FormData data = new FormData(theForm);
  HttpRequest.request(action, method: theForm.method, sendData: data)
      .then((HttpRequest req) {
    showfiles();
    SelectElement modselect = querySelector('#modselect');
    modselect.selectedIndex = 0;
  });
}

make_zipfile(Event event) {
  FormElement theForm = querySelector('#filesform');
  theForm.action = 'cgi-bin/get_zip.py';
  SelectElement downloads = querySelector("#downloads");
  if (downloads.value.length == 0) {
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
  SelectElement downloads = querySelector("#downloads");
  SelectElement modselect = querySelector('#modselect');
  if (modselect.value == SelectionModes[0]) {
    for (OptionElement o in downloads.children) {
      //select none
      o.selected = false;
    }
  } else if (modselect.value == SelectionModes[3]) {
    //select all
    for (OptionElement o in downloads.children) {
      o.selected = true;
    }
  } else if (modselect.value == SelectionModes[1]) {
    //select today
    sortOptions(downloads, datesort);
    String today = sformat(new DateTime.now());
    for (OptionElement o in downloads.children) {
      Map meta = metaParse(o.value);
      if (meta['date'] == today) {
        o.selected = true;
      } else {
        o.selected = false;
      }
    }
  } else if (modselect.value == SelectionModes[2]) {
    //select last seven days
    sortOptions(downloads, datesort);
    List lastseven = lastSevenDays();
    for (OptionElement o in downloads.children) {
      if (lastseven.contains(metaParse(o.value)['date'])) {
        o.selected = true;
      } else {
        o.selected = false;
      }
    }
  } else if (modselect.value == SelectionModes[4]) {
    sortOptions(downloads, sitesort);
  } else if (modselect.value == SelectionModes[5]) {
    for (OptionElement o in downloads.children) {
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
  for (OptionElement item in element.children) {
    myList.add(item.value);
  }
  myList.sort(compare);

  while (element.children.length > 0) {
    element.children.removeLast();
  }
  for (String item in myList) {
    OptionElement o = new OptionElement();
    o.text = item;
    o.value = item;
    element.children.add(o);
  }
}

lastSevenDays() {
  DateTime today = new DateTime.now();
  List theList = [today];
  int count = 1;
  while (theList.length < 7) {
    theList.add(today.subtract(new Duration(days: count)));
    count += 1;
  }
  List newList = [];
  for (DateTime d in theList) {
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
  InputElement t = querySelector("#siteid");

  if (int.tryParse(t.value) == null) {
    return false;
  }

  num siteid = t.valueAsNumber;
  num max = int.parse(t.attributes['max']);
  num min = int.parse(t.attributes['min']);
  if (siteid > max || siteid < min) {
    return false;
  }
  if (CurrTest == 'tmx4') {
    InputElement t = querySelector("#serial");
    if (int.tryParse(t.value) == null) {
      return false;
    }
    num serial = t.valueAsNumber;
    num max = int.parse(t.attributes['max']);
    num min = int.parse(t.attributes['min']);
    if (serial > max || serial < min) {
      return false;
    }
  }
  return true;
}

getInputs() {
  String serial = '';
  InputElement t = querySelector("#siteid");
  String siteid = t.value;

  if (CurrTest == 'tmx4') {
    InputElement s = querySelector("#serial");
    serial = s.value.padLeft(8, '0');
  }

  Map out = {'siteid': siteid, 'address': siteid, 'serial': serial};

  return out;
}

void starttest(event) {
  event.preventDefault();
  ButtonElement button = event.target;
  bool validated = validateInputs();

  if (!validated) {
    window.alert('Please correct input and try again.');
    return;
  }

  if (button.text == 'New test') {
    InputElement t = querySelector("#siteid");
    t.value = "";
    t.disabled = false;
    t.focus();
    if (CurrTest == 'tmx4') {
      InputElement serial = querySelector('#serial');
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

    DivElement z = querySelector('#messages');
    while (z.children.length > 0) {
      z.children.removeLast();
    }
    var ok = new ParagraphElement();
    ok.appendText("OK. Starting test.");
    z.append(ok);

    HttpRequest request = new HttpRequest();
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
    request.send(out);

////  HttpRequest.getString(path).then((data) {
//    HeadingElement top = querySelector("#top");
//    top.scrollIntoView();
//  });

    new Timer(new Duration(seconds: 5), checkstatus);
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
    Map data = {};
    try {
      data = jsonDecode(resp);
    } on FormatException {
      // got incomplete JSON file
      starttimer();
      return;
    }
    if (data['status'] == 'done' && data['filename'] != null) {
      doDoneStatus(data);
      return;
    } else if (data['status'] == 'new test') {
      starttimer();
      return;
    } else if (data['status'] == 'tests in progress') {
      showmessages(data['messages']);
      if (data['address'] != null) {
        if (CurrTest != '') {
          InputElement t = querySelector("#siteid");
//        t.value = "Testing - ${data['address']}";
          t.value = data['address'];
          if (t != null) {
            t.disabled = true;
          }
          if (CurrTest == 'tmx4') {
            InputElement ie = querySelector("#serial");
            if (ie != null) {
              var serial = data['serial'];
              if (serial != null && serial != '') {
                ie.value = serial;
              }
              ie.disabled = true;
            }
          }
          ButtonElement button = querySelector("#testbutton");
          if (button != null) {
            button.text = "Test in progress";
            button.disabled = true;
          }
        }
      }
      starttimer();
      return;
    }
    ButtonElement testbutton = querySelector("#testbutton");
    testbutton.text = 'Retry';
    testbutton.disabled = false;
  }).catchError(noControlFile);
}

sizeMessageBlock() {
  DivElement msgs = querySelector('#messageblock');
  Element testing = querySelector('#testing');
  Element top = querySelector('#top');
  //var clientheight = document.window.

  DivElement messageContent = querySelector('#messages');
  int height = testing.clientHeight + top.clientHeight;

  int yavailable = window.innerHeight;
  int ht = yavailable - height;
  msgs.style.top = "${height}";
  //window.alert("setting content height to ${ht}");
  //msgs.style.height = "{ht-10}px";
  messageContent.style.height = "${ht - 60}px";
}

starttimer() {
  new Timer(POLLFREQUENCY, checkstatus);
}

doDoneStatus(data) {
  String filename = data['filename'];
  downloadFile = filename;
  //InputElement addressinput = querySelector('#address');
  //addressinput.disabled = true;
  //addressinput.value = "Completed: ${data['address']}";
  show('#file_available');

  ButtonElement testbutton = querySelector("#testbutton");

  testbutton.text = "New test";
  testbutton.disabled = false;
  FormElement form = querySelector('#download_latest');
  form.method = "POST";
  form.action = 'cgi-bin/getfile.py?filename=${filename}';
  DivElement msgs = querySelector('#messages');
  sizeMessageBlock();
  if (msgs.children.length > 0) {
    var z = new ParagraphElement();
    z.appendText("done");
    msgs.append(z);
    z.scrollIntoView();
    AudioElement bell = querySelector("#bell");
    if (bell != null) {
      bell.play();
    }
  }
//  HeadingElement top = querySelector("#mainbody");
//  top.scrollIntoView();
}

getFile(event) {
  if (downloadFile != null && downloadFile.length > 3) {
    String download = 'cgi-bin/getfile.py?filename=${downloadFile}';
    HttpRequest.getString(download);
  }
}

noControlFile(event) {
  //window.alert("Warning: Control file not found. (harmless!)");
  //InputElement t = querySelector("#address");
  //t.disabled = true;
  //t.value = 'Press "New test" to begin.';
  window.alert('${event}');
  ButtonElement button = querySelector("#testbutton");
  button.disabled = false;
  starttimer();
}

showmessages(List aList) {
  DivElement liststart = querySelector('#messages');
  while (liststart.children.length > 0) {
    liststart.children.removeLast();
  }
  for (var item in aList) {
    if (item is List) {
      DateTime starttime = DateTime.parse(item[2]).toLocal();
      DateTime currtime = new DateTime.now();
      DateTime endtime = starttime.add(new Duration(seconds: item[1]));
      int diff = endtime.difference(currtime).inSeconds;
      if (diff > 0) {}
      item = item[0];
    }

    var z = new ParagraphElement();
    z.appendText(item);
    liststart.append(z);
    sizeMessageBlock();
    z.scrollIntoView();
  }
}
