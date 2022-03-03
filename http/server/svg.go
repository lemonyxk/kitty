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

var emptySVG = `<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiSvgIcon-root MuiSvgIcon-fontSizeLarge css-zjt8k" focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="CloudDownloadIcon" tabindex="-1" title="CloudDownload"></svg>`
var dirSVG = `<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiSvgIcon-root MuiSvgIcon-fontSizeLarge css-zjt8k" focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="DriveFileMoveIcon" tabindex="-1" title="DriveFileMove"><path d="M20 6h-8l-2-2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm-6 12v-3h-4v-4h4V8l5 5-5 5z"></path></svg>`
var downloadSVG = `<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiSvgIcon-root MuiSvgIcon-fontSizeLarge css-zjt8k" focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="CloudDownloadIcon" tabindex="-1" title="CloudDownload"><path d="M19.35 10.04C18.67 6.59 15.64 4 12 4 9.11 4 6.6 5.64 5.35 8.04 2.34 8.36 0 10.91 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96zM17 13l-5 5-5-5h3V9h4v4h3z"></path></svg>`
var fileSVG = `<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiSvgIcon-root MuiSvgIcon-fontSizeLarge css-zjt8k" focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="ArticleIcon" tabindex="-1" title="Article"><path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z"></path></svg>`
var backSVG = `<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiSvgIcon-root MuiSvgIcon-fontSizeLarge css-zjt8k" focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="ArrowBackIcon" tabindex="-1" title="ArrowBack"><path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"></path></svg>`
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
		.dir {
		    color: #008CBA;
		    line-height: inherit;
		    text-decoration: none;
		}
		.file {
		    color: #000000;
		    line-height: inherit;
		    text-decoration: none;
		}
		a:hover {
		    color: #008C0A;
		}
		div {
		    margin-top: 4px;
		    display: flex;
		    justify-content: flex-start;
		    align-items: center;
		}
		svg {
			width:1rem;
			margin-right: 5px;
		}
		a {
			display: flex;
			justify-content: center;
			align-items: center;
			font-size: 1.1rem;
		}
	</style>
    <title>kitty-server</title>
</head>
<body>
{{body}}
</body>

<script>

</script>

</html>
`
