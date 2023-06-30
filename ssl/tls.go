/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-26 01:56
**/

package ssl

import (
	"crypto/tls"
)

func LoadTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	return config, nil
}

func TSLConfig(certPEMBlock, keyPEMBlock []byte) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	return config, nil
}
