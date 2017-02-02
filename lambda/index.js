const spawn = require('child_process').spawn;

var AWS = require('aws-sdk');
var REGION='us-west-2';
var kms = new AWS.KMS({region:REGION});

function main() {
	var params = {
	  CiphertextBlob: Buffer('AQECAHgvkvAzsPTAcge7mEBYWxQ6m7/goE72NUHjE6hncz/4AQAAAI8wgYwGCSqGSIb3DQEHBqB/MH0CAQAweAYJKoZIhvcNAQcBMB4GCWCGSAFlAwQBLjARBAzCTqBpZtMrBuHYP9YCARCAS7M85RRcVmkT67q4ddjyPEs99MoCffMDPm9K0LV3JJ20dhg2LZRS6yanj9vBXNRBdUHD0aUD1vyPiX1xkTMM+3B6Hn9+eR22W181eQ==', 'base64'),
	};

	kms.decrypt(params, function(err, data) {
		if (err) {
			console.log(err, err.stack);
			return
		}
		var ftpsURL = data['Plaintext'].toString();

		const cmd = spawn('./runtime', {
			env: Object.assign({}, process.env, {
				S3_FTP_BUCKET: 'ftp.dev.tpt',
				TPT_POSTGRES: '',
				AWS_DEFAULT_REGION: REGION,
				AWS_SNS_TOPIC: 'arn:aws:sns:us-west-2:543266951854:tpt-alerts',
				FTP_UPLOAD: ftpsURL,
			}),
		})

		

		cmd.stdout.on('data', (data) => {
			console.log(`stdout: ${data}`);
		});

		cmd.stderr.on('data', (data) => {
			console.log(`stderr: ${data}`);
		});

		cmd.on('close', (code) => {
			console.log(`child process exited with code ${code}`);
			if (code != 0) {
				throw("Non 0 exit code")
			}
		});
	});
}

module.exports = {
	handler: main,
}
