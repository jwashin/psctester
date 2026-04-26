// ignore_for_file: unnecessary_brace_in_string_interps

// import 'dart:html';
import 'dart:js_interop';
// import 'dart:js_interop_unsafe';

import 'package:web/web.dart';
import 'package:http/http.dart' as http;
import 'dart:async';
import 'dart:convert';
import 'package:intl/intl.dart';
// import 'package:uuid/uuid.dart';

final pollFrequency = const Duration(seconds: 2);

String currentlogfilename = "";
String downloadFile = "";

// Init the global timer with nothing just so it's here.
// It only stops when we see "done". It restarts every time we checkstatus.
Timer timer = Timer(const Duration(seconds: 0), () => 'OK');

bool mainTimedOut = false;
// Uuid uuid = Uuid();

String currTest = '';

final List<String> selectionModes = [
  'select none',
  "today",
  "the last seven days",
  'select all',
  'sort by siteid',
  'invert selection',
];

void hide(String selector) {
  HTMLDivElement item = document.querySelector(selector) as HTMLDivElement;
  // item.ariaHidden = 'true';
  // item.setAttribute('hidden', 'true');
  item.classList.add('hidden');
  // item.classes.add('hidden');
}

void show(String selector) {
  HTMLDivElement item = document.querySelector(selector) as HTMLDivElement;
  // item.styles.hidden = false;
  // item.ariaHidden = 'false';
  item.classList.remove('hidden');
  // item.hidden = false;
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

  HTMLButtonElement stopButton =
      document.querySelector("#stopbutton") as HTMLButtonElement;
  stopButton.onClick.listen((e) {
    doStop();
  });

  final eventSource = EventSource('/events');

  eventSource.onMessage.listen((MessageEvent event) {
    final String rawData = event.data?.toString() ?? '';

    Map<String, dynamic> data;
    try {
      data = jsonDecode(rawData) as Map<String, dynamic>;
      // print('Received event: ${data}');
    } catch (e) {
      print('Error decoding JSON: $e');
      print('Raw data: ${rawData}');
      return;
    }
    hide('#testinput');
    show('#testing');
    // Extract your LogWriter fields
    final String status = data['status'] ?? 'No Status';
    final String? filename = data['filename'];
    final List<dynamic> messages = data['messages'] ?? [];
    updateDashboard(status, messages);
    if (status == 'done' && filename != null) {
      // eventSource.close(); // Stop the stream
      doDoneStatus(data);
    }
  });

  // 3. Optional: Handle connection state
  eventSource.onOpen.listen((_) => print('Connected to SSE Stream'));
  eventSource.onError.listen((_) => print('SSE Connection Error/Lost'));

  HTMLButtonElement tmx5button =
      document.querySelector("#test_tmx5") as HTMLButtonElement;
  // tmx5button.
  tmx5button.onClick.listen((e) {
    doTmx5(e);
  });

  HTMLButtonElement tmx5nbutton =
      document.querySelector("#test_tmx5n") as HTMLButtonElement;
  tmx5nbutton.onClick.listen((e) {
    doTmx5n(e);
  });

  HTMLButtonElement qfambutton =
      document.querySelector("#test_Qfam") as HTMLButtonElement;
  qfambutton.onClick.listen((e) {
    doQfam(e);
  });

  HTMLButtonElement tmx4button =
      document.querySelector("#test_tmx4") as HTMLButtonElement;
  tmx4button.onClick.listen((e) {
    doTmx4(e);
  });

  HTMLButtonElement tmx1button =
      document.querySelector("#test_tmx1") as HTMLButtonElement;
  tmx1button.onClick.listen((e) {
    doTmx1(e);
  });

  HTMLButtonElement tmx3button =
      document.querySelector("#test_tmx3") as HTMLButtonElement;
  tmx3button.onClick.listen((e) {
    doTmx3(e);
  });

  HTMLButtonElement files =
      document.querySelector("#manage_files") as HTMLButtonElement;

  files.onClick.listen((e) {
    doFiles(e);
  });

  //testbutton.disabled = true;

  HTMLButtonElement maintenance =
      document.querySelector("#maintenance") as HTMLButtonElement;

  maintenance.onClick.listen((e) {
    doMaintenance(e);
  });

  //  ButtonElement filesbutton = document.querySelector("#downloads_button");
  // filesbutton.onClick.listen(showfiles);

  // TODO: the next block contains the nost recent version of this

  // window.addEventListener(
  //   'onResize',
  //   (e) {
  //     Element? z = document.querySelector('#messageblock');
  //     if (z != null && z.getAttribute('visible') == 'true') {
  //       sizeMessageBlock();
  //     }
  //   }.toJS,
  // );

  // window.onResize.listen((e) {
  //   Element? z = document.querySelector('#messageblock');
  //   if (z != null && z.hidden == false) {
  //     sizeMessageBlock();
  //   }
  // });

  // checkstatus();
}

void doFiles(Event event) {
  hide('#main');
  show('#files');
  HTMLButtonElement returnbutton =
      document.querySelector("#return_main2") as HTMLButtonElement;
  returnbutton.onClick.listen((e) {
    returnToMain(e);
  });
  showfiles();
  HTMLSelectElement modselect =
      document.querySelector('#modselect') as HTMLSelectElement;
  modselect.value = selectionModes[0];
  modselect.onChange.listen((e) {
    doFileSelection();
  });

  HTMLButtonElement download =
      document.querySelector('#download') as HTMLButtonElement;
  download.onClick.listen((Event e) {
    downloadCombinedCsv(e);
  });
  HTMLButtonElement archive =
      document.querySelector('#archive') as HTMLButtonElement;
  archive.onClick.listen((Event e) {
    archiveFiles(e);
  });
}

void doTmx4(Event event) {
  currTest = 'tmx4';
  show("#testing");
  show('#serial_entry');
  // Element testing = document.querySelector("#testing")!;
  // testing.classes.remove('hidden');
  enableTest();
  //  show("#siteidinput");

  //InputElement siteid = document.querySelector("#siteid");
  hide('#main');
  rtMainEnable();
}

void doTmx1(Event event) {
  currTest = 'tmx1';
  show("#testinput");
  show('#serial_entry');
  // Element testing = document.querySelector("#testing")!;
  // testing.classes.remove('hidden');
  enableTest();
  //  show("#siteidinput");

  //InputElement siteid = document.querySelector("#siteid");
  hide('#main');
  rtMainEnable();
}

void doTmx3(Event event) {
  currTest = 'tmx3';
  show("#testinput");
  show('#serial_entry');
  show('#tmx3_type');
  // Element testing = document.querySelector("#testing")!;
  // testing.classes.remove('hidden');
  enableTest();
  //  show("#siteidinput");

  //InputElement siteid = document.querySelector("#siteid");
  hide('#main');
  rtMainEnable();
}

void enableTest() {
  HTMLButtonElement testbutton =
      document.querySelector('#testbutton') as HTMLButtonElement;
  testbutton.textContent = "Begin test";
  testbutton.disabled = false;
  show('#messageblock');
  HTMLInputElement siteid =
      document.querySelector('#siteid') as HTMLInputElement;
  siteid.disabled = false;
  siteid.focus();
  sizeMessageBlock();
  testbutton.onClick.listen((e) {
    starttest(e);
  });
}

void doTmx5(Event event) {
  show("#testinput");
  hide("#serial_entry");
  // Element s = document.querySelector()!;
  // s.classes.add('hidden');
  // s.hidden = true;

  currTest = 'tmx5';
  // Element testing = document.querySelector("#testing")!;
  // testing.classes.remove('hidden');
  enableTest();
  HTMLInputElement serial =
      document.querySelector("#serial") as HTMLInputElement;
  serial.value = '';
  hide("#main");
  rtMainEnable();
}

void doTmx5n(Event event) {
  show("#testinput");
  hide("#serial_entry");
  // Element s = document.querySelector()!;
  // s.classes.add('hidden');
  // s.hidden = true;

  currTest = 'tmx5n';
  // Element testing = document.querySelector("#testing")!;
  // testing.classes.remove('hidden');
  enableTest();
  HTMLInputElement serial =
      document.querySelector("#serial") as HTMLInputElement;
  serial.value = '';
  hide("#main");
  rtMainEnable();
}

void doQfam(Event event) {
  show("#testinput");
  hide("#serial_entry");
  // Element s = document.querySelector()!;
  // s.classes.add('hidden');
  // s.hidden = true;

  currTest = 'Qfam';
  // Element testing = document.querySelector("#testing")!;
  // testing.classes.remove('hidden');
  enableTest();
  HTMLInputElement serial =
      document.querySelector("#serial") as HTMLInputElement;
  serial.value = '';
  hide("#main");
  rtMainEnable();
}

void rtMainEnable() {
  HTMLButtonElement returnbutton =
      document.querySelector("#return_main") as HTMLButtonElement;
  returnbutton.onClick.listen((e) {
    returnToMain(e);
  });
}

void doMaintenance(Event event) {
  hide("#main");
  hide('#messageblock');
  var maint = document.querySelector("#sysmaint") as Element;
  maint.classList.remove('hidden');
  show('#sysmaint');
  showdate('#dt1');
  getVersions();
  getAddresses();
  var settime = document.querySelector("#time_fix") as HTMLButtonElement;
  settime.onClick.listen((e) {
    fixTime(e);
  });

  HTMLButtonElement returnbutton =
      document.querySelector("#go_main") as HTMLButtonElement;
  returnbutton.onClick.listen((e) {
    returnToMain(e);
  });

  HTMLButtonElement reinitbutton =
      document.querySelector("#reinit") as HTMLButtonElement;
  reinitbutton.onClick.listen((e) async {
    Uri path = Uri.parse('/cgi-bin/initcontrolfile.py');

    await http.get(path);
    window.alert('Control file reinitialized!');
  });
}

void getVersions() {
  Uri uri = Uri.parse("/script_ver");
  HTMLPreElement pretag = document.querySelector('#versions') as HTMLPreElement;
  http.get(uri).then((var resp) {
    pretag.textContent = resp.body;
  });

  Element? s = document.querySelector('#swver');
  uri = Uri.parse("/version_tag.txt");
  http.get(uri).then((var resp) {
    s?.textContent = "Software versions: ( ${resp.body} )";
  });
}

void getAddresses() {
  HTMLSpanElement p = document.querySelector("#ip") as HTMLSpanElement;
  while ((p.firstElementChild != null)) {
    p.firstElementChild?.remove();
  }
  Uri uri = Uri.parse("/cgi-bin/ip_addr.py");

  http.get(uri).then((var resp) {
    Map j = jsonDecode(resp.body);
    for (String item in j.keys as Iterable<String>) {
      String? address = j[item];
      if (item[0] == 'w') {
        p.textContent = '${p.textContent} Wireless: $address ';
      } else {
        p.textContent = '${p.textContent} Wired: $address ';
      }
      p.children.add(HTMLBRElement());
    }
  });
}

Future<void> fixTime(Event event) async {
  DateTime t = DateTime.now().toUtc();
  List s = [t.year, t.month, t.day, t.hour, t.minute, t.second, t.millisecond];

  Uri path = Uri.parse('/cgi-bin/settime.py');

  String out = jsonEncode({'dt': s});
  try {
    var response = await http.post(
      path,
      body: out,
      headers: {'Content-Type': 'application/json'},
    );
    if (response.statusCode == 200) {
      // handle response similarly to onTimeFix
      try {
        Map j = jsonDecode(response.body);
        if (j['resp'] == true) {
          showdate('#dt1');
          window.alert('System time updated to ${t.toLocal()}!');
        } else {
          window.alert('time update failed!');
        }
      } on FormatException {
        window.alert('Invalid response from server when setting time.');
      }
    } else {
      window.alert('Error setting time: ${response.statusCode}');
    }
  } catch (e) {
    // network or other error
    mainTimeout();
  }
}

// void onTimeFix(HttpRequest request, DateTime t) {
//   String resp = request.responseText!;
//   Map j = jsonDecode(resp);
//   if (j['resp'] == true) {
//     showdate('#dt1');
//     window.alert('System time updated to ${t.toLocal()}!');
//   } else {
//     window.alert('time update failed!');
//   }
// }

void returnToMain(Event event) {
  window.location.href = '/';
}

void clearFiles() {
  HTMLSelectElement downloads =
      document.querySelector("#downloads") as HTMLSelectElement;
  while (downloads.firstElementChild != null) {
    downloads.firstElementChild?.remove();
  }
}

void showdate(String loc) {
  Uri path = Uri.parse('/cgi-bin/getdate.py');
  HTMLSpanElement? dt = document.querySelector(loc) as HTMLSpanElement?;
  http.get(path).then((var resp) {
    String dtstring = jsonDecode(resp.body)['datetime'];
    DateTime value = DateTime.parse(dtstring).toLocal();
    String txt = DateFormat.yMEd().add_jms().format(value);
    dt!.textContent = txt;
  });
}

Future<void> showfiles() async {
  Uri path = Uri.parse('cgi-bin/fileslist.py');
  HTMLSelectElement modselect =
      document.querySelector('#modselect') as HTMLSelectElement;

  if (modselect.options.length == 0) {
    for (var mode in selectionModes) {
      HTMLOptionElement o = HTMLOptionElement();
      o.text = mode;
      o.value = mode;
      modselect.options.add(o);
    }
  }

  clearFiles();

  HTMLSelectElement downloads =
      document.querySelector("#downloads") as HTMLSelectElement;

  var resp = await http.get(path);

  var data = jsonDecode(resp.body)['files'];

  for (String filename in data) {
    HTMLOptionElement o = HTMLOptionElement();
    o.text = filename;
    o.value = filename;
    downloads.options.add(o);

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

  //  SelectElement modselect = document.querySelector('#modselect');
  show('#modselect');
}

Future<void> archiveFiles(Event event) async {
  event.preventDefault();
  event.stopPropagation();
  HTMLFormElement theForm =
      document.querySelector('#filesform') as HTMLFormElement;
  Uri action = Uri.parse('/cgi-bin/archive_files.py');
  HTMLSelectElement downloads =
      document.querySelector("#downloads") as HTMLSelectElement;
  int count = downloads.selectedOptions.length;
  if (count >= 1) {
    bool resp = window.confirm('Moving $count files to Archive. Continue?');
    if (resp == false) {
      return;
    }
    FormData data = FormData(theForm);

    await http.post(action, body: data);
    showfiles();
    HTMLSelectElement modselect =
        document.querySelector('#modselect') as HTMLSelectElement;
    modselect.selectedIndex = 0;
  }
}

void downloadCombinedCsv(Event event) {
  event.preventDefault();

  final downloads = document.querySelector("#downloads") as HTMLSelectElement;
  final files = <String>[];

  final options = downloads.options;
  for (int i = 0; i < options.length; i++) {
    final option = options.item(i) as HTMLOptionElement;
    if (option.selected) {
      files.add(option.value);
    }
  }
  if (files.isEmpty) {
    event.preventDefault();
    event.stopPropagation();
    window.alert('Please select files for download,');
    return;
  }

  // Create a hidden form to 'POST' the data
  final form = document.createElement('form') as HTMLFormElement;
  form.method = 'POST';
  form.action = '/cgi-bin/get_combined.py';

  final input = document.createElement('input') as HTMLInputElement;
  input.type = 'hidden';
  input.name = 'files';
  input.value = files.join(',');

  form.append(input);
  document.body!.append(form);
  form.submit();
  form.remove();
}

void makeZipfile(Event event) async {
  event.preventDefault();
  // HTMLFormElement theForm =
  // document.querySelector('#filesform') as HTMLFormElement;
  // theForm.action = '/cgi-bin/get_zip.py';

  HTMLSelectElement downloads =
      document.querySelector("#downloads") as HTMLSelectElement;
  List<String> files = [];

  final options = downloads.options;
  for (int i = 0; i < options.length; i++) {
    final option = options.item(i) as HTMLOptionElement;
    if (option.selected) {
      files.add(option.value);
    }
  }
  if (files.isEmpty) {
    event.preventDefault();
    event.stopPropagation();
    window.alert('Please select files for download,');
    return;
  }

  Uri path = Uri.parse('/cgi-bin/get_zip.py');

  try {
    final response = await http.post(path, body: {'files': files.join(',')});
    window.alert('response received, status code: ${response.statusCode}');

    if (response.statusCode == 200) {
      // 2. Wrap the bytes in a Blob
      final blob = Blob(
        [response.bodyBytes.toJS].toJS,
        BlobPropertyBag(type: 'application/zip'),
      );
      final url = URL.createObjectURL(blob);

      // 3. Trigger the browser save dialog
      final anchor = document.createElement('a') as HTMLAnchorElement;
      anchor.href = url;
      anchor.target = '_blank';
      // acc_2026-04-25_13-00.zip
      String date = DateTime.now().toIso8601String().replaceAll(':', '-');
      String filename = 'acc_${date}.zip';
      anchor.download = filename;

      // Append, click, and remove to ensure browser compatibility
      document.body!.append(anchor);
      anchor.click();
      anchor.remove();

      // 5. Cleanup memory
      URL.revokeObjectURL(url);
    } else {
      // Handle error (e.g., show an alert on the phone)
      window.alert('Download failed: ${response.statusCode}');
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
  } catch (e) {
    console.error('Full Error: $e'.toJS);
    window.alert('Error caught: $e');
  }
}

void doFileSelection() {
  //List SelectionModes = ['select none', "today's files", "this week's files", 'select all'];
  HTMLSelectElement downloads =
      document.querySelector('#downloads') as HTMLSelectElement;
  HTMLSelectElement modselect =
      document.querySelector('#modselect') as HTMLSelectElement;
  // just to be sure
  downloads.multiple = true;
  if (modselect.value == selectionModes[0]) {
    // select none
    downloads.selectedIndex = -1;
  } else if (modselect.value == selectionModes[3]) {
    //select all

    for (int i = 0; i < downloads.options.length; i++) {
      var o = downloads.options.item(i) as HTMLOptionElement;
      o.selected = true;
    }
  } else if (modselect.value == selectionModes[1]) {
    //select today
    sortOptions(downloads, datesort);
    String today = sformat(DateTime.now());
    for (int i = 0; i < downloads.options.length; i++) {
      var o = downloads.options.item(i) as HTMLOptionElement;
      Map meta = metaParse(o.value);
      if (meta['date'] == today) {
        o.selected = true;
      } else {
        o.selected = false;
      }
    }
  } else if (modselect.value == selectionModes[2]) {
    //select last seven days
    sortOptions(downloads, datesort);
    List lastseven = lastSevenDays();
    for (int i = 0; i < downloads.options.length; i++) {
      var o = downloads.options.item(i) as HTMLOptionElement;
      if (lastseven.contains(metaParse(o.value)['date'])) {
        o.selected = true;
      } else {
        o.selected = false;
      }
    }
  } else if (modselect.value == selectionModes[4]) {
    sortOptions(downloads, sitesort);
  } else if (modselect.value == selectionModes[5]) {
    for (int i = 0; i < downloads.options.length; i++) {
      var o = downloads.options.item(i) as HTMLOptionElement;
      if (o.selected == true) {
        o.selected = false;
      } else {
        o.selected = true;
      }
    }
  }
}

void sortOptions(HTMLSelectElement element, Comparator compare) {
  var myList = [];
  for (int i = 0; i < element.options.length; i++) {
    HTMLOptionElement item = element.options.item(i) as HTMLOptionElement;
    myList.add(item.value);
  }
  myList.sort(compare);

  while (element.firstElementChild != null) {
    element.firstElementChild?.remove();
  }
  for (var item in myList) {
    HTMLOptionElement o = HTMLOptionElement();
    o.text = item;
    o.value = item;
    element.options.add(o);
  }
}

List lastSevenDays() {
  DateTime today = DateTime.now();
  List theList = [today];
  int count = 1;
  while (theList.length < 7) {
    theList.add(today.subtract(Duration(days: count)));
    count += 1;
  }
  List newList = [];
  for (var d in theList) {
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

Map metaParse(String s) {
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

String sformat(DateTime adate) {
  String month = '${adate.month}'.padLeft(2, '0');
  String day = '${adate.day}'.padLeft(2, '0');
  String stoday = "${adate.year}${month}${day}";
  return stoday;
}

bool validateInputs() {
  // site id must be numeric and between max and min
  HTMLInputElement t = document.querySelector("#siteid") as HTMLInputElement;

  int? max = int.tryParse(t.max);
  int? min = int.tryParse(t.min);

  if (int.tryParse(t.value) == null) {
    return false;
  }

  int siteid = (t.valueAsNumber).toInt();
  if ((max == null || min == null) || (siteid > max || siteid < min)) {
    return false;
  }
  if (['tmx4', 'tmx3', 'tmx1'].contains(currTest)) {
    HTMLInputElement t = document.querySelector("#serial") as HTMLInputElement;
    int? max = int.tryParse(t.max);
    int? min = int.tryParse(t.min);

    if (int.tryParse(t.value) == null) {
      return false;
    }
    num serial = t.valueAsNumber;

    if ((max == null || min == null) || (serial > max || serial < min)) {
      return false;
    }
  }
  return true;
}

Map getInputs() {
  String serial = '';
  HTMLInputElement t = document.querySelector("#siteid") as HTMLInputElement;
  String? siteid = t.value;

  // address is included for historical reasons
  Map out = {'siteid': siteid, 'address': siteid, 'c': currTest};

  if (['tmx4', 'tmx3', 'tmx1'].contains(currTest)) {
    HTMLInputElement s = document.querySelector("#serial") as HTMLInputElement;
    serial = s.value.padLeft(8, '0');
    out['serial'] = serial;
  }
  if (currTest == 'tmx3') {
    String x3 = "";
    HTMLInputElement rad =
        document.querySelector('input[name="tmx3"]:checked')
            as HTMLInputElement;
    x3 = rad.value;
    // RadioNodeList radios =
    //     document.querySelectorAll('input[name="tmx3"]') as RadioNodeList;
    // x3 = radios.value;

    Element channelInput = document.querySelector("#channel_input")!;
    channelInput = channelInput as HTMLInputElement;
    String v = "";
    var chv = [];
    v = channelInput.value;
    var s = v.split("-");
    for (var i in s) {
      chv.add(int.parse(i));
    }
    if (chv.isNotEmpty) {
      out['channels'] = chv;
    }
    out['x3'] = x3;
  }

  return out;
}

void starttest(Event event) async {
  event.preventDefault();
  HTMLButtonElement button = event.target as HTMLButtonElement;
  bool validated = validateInputs();

  if (!validated) {
    window.alert('Please correct input and try again.');
    return;
  }

  if (button.textContent == 'New test') {
    HTMLInputElement t = document.querySelector("#siteid") as HTMLInputElement;
    t.value = "";
    t.disabled = false;
    t.focus();
    if (currTest == 'tmx4') {
      HTMLInputElement serial =
          document.querySelector('#serial') as HTMLInputElement;
      serial.disabled = false;
      serial.value = '';
    }
    button.disabled = false;
    button.textContent = 'Begin test';
  } else {
    button.textContent = "Test in progress";
    button.disabled = true;
    sizeMessageBlock();

    hide('#file_available');
    show('#testing');
    HTMLDivElement messageBlock =
        document.querySelector('#messages') as HTMLDivElement;

    while (messageBlock.firstElementChild != null) {
      messageBlock.firstElementChild?.remove();
    }
    var ok = HTMLParagraphElement();
    ok.textContent = "OK. Starting test.";
    messageBlock.append(ok);

    // HttpRequest request = HttpRequest();
    // request.onLoadEnd.listen((e) {
    //   onResponse(request);
    // });
    // request.onTimeout.listen((e) {
    //   mainTimeout();
    // });
    Uri path = Uri.parse('/cgi-bin/dotest.py');

    // request.open('POST', path);
    //String address = getSiteId();

    String out = jsonEncode(getInputs());

    var response = await http.post(
      path,
      headers: {'Content-Type': 'application/json'},
      body: out,
    );
    hide('#testinput');
    if (response.statusCode != 200) {
      window.alert('Error starting test: ${response.statusCode}');
    }
  }
}

void mainTimeout() {
  mainTimedOut = true;
  window.alert("main script timed out (harmless)");
  //  Location loc =window.location;
  //  document.window.location = loc;
}

// // this is the return from starting a test
// // it doesn't have to do anything, since test status is in control.json
// void onResponse(Response response) {}

// put the display in "test in progress" mode
// if status is "tests in progress" at startup
// void startStatusStream() {
//   // Use a relative path. The browser prepends the current IP/Port automatically.

//   source.onmessage = (MessageEvent event) {
//     final String rawData = (event.data as JSString).toDart;
//     console.log(("Received event: ${rawData}").toJS);

//     if (rawData.isEmpty) {
//       return; // Skip if Go is still writing/file is empty
//     }
//     if (rawData.contains("heartbeat")) {
//       return;
//     }

//     // 1. Decode the JSON
//     Map data;
//     try {
//       data = jsonDecode(rawData);
//     } catch (e) {
//       window.alert("Received invalid JSON data: ${rawData}");
//       return; // Skip if Go is still writing/file is empty
//     }

//     // 2. Handle "Done" State
//     if (data['status'] == 'done' && data['filename'] != null) {
//       source.close(); // Stop the stream
//       doDoneStatus(data);
//       return;
//     }

//     // 3. Handle "In Progress" State
//     if (data['status'] == 'tests in progress') {
//       document.querySelector('#testing')?.classList.remove('hidden');
//       document.querySelector('#main')?.classList.add('hidden');

//       // Update messages
//       if (data['messages'] != null) {
//         showmessages(data['messages']);
//       }

//       // Handle UI Updates (replacing the old HTMLInputElement casts)
//       if (data['address'] != null) {
//         final siteId = document.querySelector('#siteid') as HTMLInputElement;
//         siteId.value = data['address'];
//         siteId.disabled = true;

//         document.querySelector('#messageblock')?.classList.remove('hidden');
//         rtMainEnable();
//       }

//       final button =
//           document.querySelector("#testbutton") as HTMLButtonElement?;
//       if (button != null) {
//         button.textContent = "Test in progress";
//         button.disabled = true;
//       }
//     }
//   }.toJS;

//   source.onerror = (Event e) {
//     // No need to call starttimer(). EventSource auto-reconnects!
//     print("Stream connection lost. Reconnecting...");
//   }.toJS;
// }

void updateDashboard(String status, List messages) {
  if (status == 'tests in progress') {
    document.querySelector('#testing')?.classList.remove('hidden');
    document.querySelector('#main')?.classList.add('hidden');
  }
  if (messages.isNotEmpty) {
    showmessages(messages);
  }
}

void sizeMessageBlock() {
  HTMLDivElement msgs =
      document.querySelector('#messageblock') as HTMLDivElement;
  Element testing = document.querySelector('#testing')!;
  Element top = document.querySelector('#top')!;
  //var clientheight = document.window.

  HTMLDivElement messageContent =
      document.querySelector('#messages') as HTMLDivElement;
  int height = testing.clientHeight + top.clientHeight;

  int yavailable = window.innerHeight;
  int ht = yavailable - height;
  msgs.style.top = "${height}";
  //window.alert("setting content height to ${ht}");
  //msgs.style.height = "{ht-10}px";
  messageContent.style.height = "${ht - 60}px";
}

// void starttimer() {
//   if (!timer.isActive) {
//     timer = Timer(pollFrequency, checkstatus);
//   }
// }

void doDoneStatus(Map data) {
  timer.cancel();
  String? filename = data['filename'];
  if (filename != null) {
    downloadFile = filename;
  }
  //InputElement addressinput = document.querySelector('#address');
  //addressinput.disabled = true;
  //addressinput.value = "Completed: ${data['address']}";
  show('#file_available');

  HTMLButtonElement testbutton =
      document.querySelector("#testbutton") as HTMLButtonElement;

  testbutton.textContent = "New test";
  testbutton.disabled = false;

  HTMLFormElement form =
      document.querySelector('#download_latest') as HTMLFormElement;
  form.method = "POST";
  form.action = '/cgi-bin/getfile.py?filename=${filename}';

  HTMLDivElement msgs = document.querySelector('#messages') as HTMLDivElement;
  sizeMessageBlock();
  if (msgs.firstElementChild != null) {
    var z = HTMLParagraphElement();
    z.textContent = "done";
    msgs.append(z);
    z.scrollIntoView();
    HTMLAudioElement? bell =
        document.querySelector("#bell") as HTMLAudioElement?;
    if (bell != null) {
      bell.play();
    }
  }
  //  HeadingElement top = document.querySelector("#mainbody");
  //  top.scrollIntoView();
}

// void getFile(Event event) {
//   if (downloadFile.length > 3) {
//     Uri download = Uri.parse('/cgi-bin/getfile.py?filename=${downloadFile}');

//     // String download = '/cgi-bin/getfile.py?filename=${downloadFile}';
//     http.get(download);
//   }
// }

// void noControlFile(Event event) {
//   //window.alert("Warning: Control file not found. (harmless!)");
//   //InputElement t = document.querySelector("#address");
//   //t.disabled = true;
//   //t.value = 'Press "New test" to begin.';
//   window.alert('${event}');
//   HTMLButtonElement button =
//       document.querySelector("#testbutton") as HTMLButtonElement;
//   button.disabled = false;
//   // starttimer();
// }

void showmessages(List aList) {
  HTMLDivElement liststart =
      document.querySelector('#messages') as HTMLDivElement;
  while (liststart.firstElementChild != null) {
    liststart.firstElementChild?.remove();
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

    var z = HTMLParagraphElement();
    z.textContent = item;
    liststart.append(z);
    sizeMessageBlock();
    z.scrollIntoView();
  }
}

Future<void> doStop() async {
  var url = Uri.parse('/stop');
  var response = await http.post(url);
  if (response.statusCode == 200) {
    print('Stop signal received by Go.');
  }
}
