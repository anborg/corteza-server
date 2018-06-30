#!/usr/bin/env php
<?php

error_reporting(E_ALL^E_NOTICE);

include("docs/vendor/autoload.php");

function array_change_key_case_recursive($arr) {
	return array_map(function ($item) {
		if (is_array($item)) {
			$item = array_change_key_case_recursive($item);
		}
		return $item;
	}, array_change_key_case($arr));
}

$tpl = new Monotek\MiniTPL\Template;
$tpl->set_compile_location("/tmp", true);
$tpl->add_default("newline", "\n");

$api_files = glob("docs/src/spec/*.json");
$apis = array_map(function($filename) {
	return array_change_key_case_recursive(json_decode(file_get_contents($filename), true));
}, $api_files);

usort($apis, function($a, $b) {
	return strcmp($a['interface'], $b['interface']);
});

foreach (array("structs", "handlers", "interfaces", "request", "") as $type) {
	foreach ($apis as $api) {
		if (is_array($api['struct'])) {
			$name = ucfirst($api['interface']);
			$filename = str_replace("..", ".", strtolower($name) . "." . $type . ".go");

			$tpl->load("http_$type.tpl");
			$tpl->assign("parsers", array(
				"uint64" => "parseUInt64"
			));
			$tpl->assign("package", $api['package']);
			$tpl->assign("name", $name);
			$tpl->assign("self", strtolower(substr($name, 0, 1)));
			$tpl->assign("api", $api);
			$tpl->assign("fields", $api['struct']);
			$tpl->assign("calls", $api['apis']);
			$contents = $tpl->get();

			$save = true;
			if ($type === "" && file_exists($filename)) {
				$save = false;
			}
			if ($save) {
				file_put_contents($filename, $contents);
			}
		}
	}
}

foreach (array("routes") as $type) {
	$name = ucfirst($api['interface']);
	$filename = str_replace("..", ".", $type . ".go");

	$tpl->load("http_$type.tpl");
	$tpl->assign("package", reset($apis)['package']);
	$tpl->assign("apis", $apis);
	$contents = $tpl->get();

	file_put_contents($filename, $contents);
}

passthru("go fmt");