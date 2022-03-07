/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2022-03-03 05:01
**/

package server

var emptySVG = `<svg class="empty-svg svg" focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="CloudDownloadIcon" tabindex="-1" title="CloudDownload"></svg>`
var dirSVG = `<svg class="dir-svg svg" focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="DriveFileMoveIcon" tabindex="-1" title="DriveFileMove"><path d="M20 6h-8l-2-2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm-6 12v-3h-4v-4h4V8l5 5-5 5z"></path></svg>`
var downloadSVG = `<svg class="download-svg svg" focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="CloudDownloadIcon" tabindex="-1" title="CloudDownload"><path d="M19.35 10.04C18.67 6.59 15.64 4 12 4 9.11 4 6.6 5.64 5.35 8.04 2.34 8.36 0 10.91 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96zM17 13l-5 5-5-5h3V9h4v4h3z"></path></svg>`
var fileSVG = `<svg class="file-svg svg" focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="ArticleIcon" tabindex="-1" title="Article"><path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z"></path></svg>`
var backSVG = `<svg class="back-svg svg" focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="ArrowBackIcon" tabindex="-1" title="ArrowBack"><path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"></path></svg>`

var html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width,initial-scale=1.0,user-scalable=0,minimum-scale=1.0,maximum-scale=1.0">
	<style>
		body {
		    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue",
		    sans-serif;
		    -webkit-font-smoothing: antialiased;
		    -moz-osx-font-smoothing: grayscale;
		}
		
		* {
		    box-sizing: border-box;
		}
		
		
		pre:hover {
		    background-color: #f3f3f3;
		}
		
		pre {
		    white-space: pre-wrap;
		    white-space: -moz-pre-wrap;
		    white-space: -pre-wrap;
		    white-space: -o-pre-wrap;
		    word-wrap: break-word;
		    background-color: #f8f8f8;
		    border: 1px solid #dfdfdf;
		    margin-top: 1.5em;
		    margin-bottom: 1.5em;
		    padding: 1rem;
		    position: relative;
		}
		
		pre code {
		    background-color: transparent;
		    border: 0;
		    padding: 0;
		}
		
		code {
		    display: block;
		    overflow-x: auto;
		    /*padding: 1em;*/
		    /*background: #f3f3f3;*/
		    /*color: #444;*/
		    background-color: #f8f8f8;
		    border-color: #dfdfdf;
		    border-style: solid;
		    border-width: 1px;
		    color: #333;
		    font-family: Consolas, "Liberation Mono", Courier, monospace;
		    font-weight: normal;
		    padding: 0.125rem 0.3125rem 0.0625rem;
		}
		
		code {
		    font-family: source-code-pro, Menlo, Monaco, Consolas, "Courier New", monospace;
		}
		
		::-webkit-scrollbar {
		    width: 0px;
		    height: 0px;
		}
		
		::-webkit-scrollbar-track {
		    /* border-radius: 3px; */
		    /* background: rgba(115, 18, 226, 0.16); */
		    box-shadow: inset 0 0 5px rgba(0, 0, 0, 0.08);
		}
		
		::-webkit-scrollbar-thumb {
		    /* border-radius: 3px; */
		    background: rgba(20, 5, 107, 0.06);
		    box-shadow: inset 0 0 5px rgba(0, 0, 0, 0.08);
		}
		
		
		table {
		    background: #fff;
		    border: solid 1px #ddd;
		    margin-bottom: 1.25rem;
		    table-layout: auto;
		}
		
		table thead {
		    background: #F5F5F5;
		}
		
		table thead tr th, table tfoot tr th, table tfoot tr td, table tbody tr th, table tbody tr td, table tr td {
		    display: table-cell;
		    line-height: 1.125rem;
		}
		
		table thead tr th, table thead tr td {
		    color: #222;
		    font-size: 0.875rem;
		    font-weight: bold;
		    padding: 0.5rem 0.625rem 0.625rem;
		}
		
		table tr th, table tr td {
		    color: #222;
		    font-size: 0.875rem;
		    padding: 0.5625rem 0.625rem;
		    text-align: left;
		}
		
		table tr.even, table tr.alt, table tr:nth-of-type(even) {
		    background: #F9F9F9;
		}
		
		img {
		    display: inline-block;
		    vertical-align: middle;
		}
		
		img {
		    -ms-interpolation-mode: bicubic;
		}
		
		img {
		    max-width: 100%;
		    height: auto;
		}
		
		.dir, .file {
		    color: #008CBA;
		    line-height: inherit;
		    text-decoration: none;
		
		    display: flex;
		    justify-content: center;
		    align-items: center;
		    font-size: 1.1rem;
		}
		
		.svg {
		    width: 1rem;
		    margin-right: 5px;
		    cursor: pointer;
		}
		
		.svg:hover {
		    transform: scale(1.1);
		}
		
		.file {
		    color: #000000;
		    line-height: inherit;
		    text-decoration: none;
		}
		
		.dir:hover, .file:hover {
		    color: #008C0A;
		}
		
		.list {
		    margin-bottom: 4px;
		    display: flex;
		    justify-content: flex-start;
		    align-items: center;
		}
		
		blockquote, blockquote p {
		    line-height: 1.6;
		    color: #6f6f6f;
		}
		
		blockquote {
		    margin: 0 0 1.25rem;
		    padding: 0.5625rem 1.25rem 0 1.1875rem;
		    border-left: 1px solid #ddd;
		}
		
		ul {
		    padding: 0;
		    margin-left: 1.1rem;
		}
		
		ul, ol, dl {
		    font-family: inherit;
		    font-size: 1rem;
		    line-height: 1.6;
		    list-style-position: outside;
		    margin-bottom: 1.25rem;
		}
		
		hr {
		    border: solid #ddd;
		    border-width: 1px 0 0;
		    clear: both;
		    height: 0;
		    margin: 1.25rem 0 1.1875rem;
		}
		
		h1, h2, h3, h4, h5, h6 {
		    font-family: 'Old Standard TT', serif;
		    font-weight: bold;
		}
		
		.copy-container {
		    width: 24px;
		    height: 24px;
		    position: absolute;
		    right: 1rem;
		    top: 1rem;
		    display: none;
		}
		
		.copy {
		    width: 24px;
		    height: 24px;
		    cursor: pointer;
		}
		
		.copy:hover {
		    transform: scale(1.1);
		}
		
		.ok {
		    width: 24px;
		    height: 24px;
		    cursor: pointer;
		    fill: green;
		}
		
		.ok:hover {
		    transform: scale(1.1);
		}
		
		svg:focus {
		    outline: none;
		}
		
		.back {
		    width: 1.5rem;
		    height: 1.5rem;
		    cursor: pointer;
		    /*fill: #008CBA;*/
		    position: fixed;
		    right: 0.5rem;
		    top: 0.5rem;
		    z-index: 999;
		}
		
		.back:hover {
		    transform: scale(1.1);
		}
	</style>
    <title>kitty-server</title>
</head>
<body>
{{body}}
</body>

<script>
    let backSVG = '<svg focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="ArrowBackIcon" tabindex="-1" title="ArrowBack"><path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"></path></svg>'

    let backDiv = document.createElement("div")

    backDiv.classList.add("back")

    document.body.appendChild(backDiv).innerHTML = backSVG;

    backDiv.addEventListener("click", function (e) {
        var url = window.location.pathname
		if (url.lastIndexOf("/") == 0) {
			window.location.href = "/"
		    return
		}
        var dir = url.substring(0, url.lastIndexOf("/"))
        window.location.href = dir
    })
</script>

</html>
`
