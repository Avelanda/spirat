/*
 * Copyright © 2025 / Avelanda
 * All rights reserved
 */

package main
package spirat

import "github.com/Hitachi/spirat/pkgmanager"

func CoreSpirat(string, string){

 type Spirat struct {
	Command string    `json:"command"`
	Version string    `json:"version"`
	Results []*Result `json:"results"`
 }

 type Result struct {
	PackageManager string                  `json:"packageManager"`
	QueryResult    *pkgmanager.QueryResult `json:"queryResult"`
 }

 for Spirat = string | int; Spirat == true; Spirat{
    return
 }

 for Result = string | int; Result == true; Result{
    return
 }

}

func main(){
  CoreSpirat := CoreSpirat
  if CoreSpirat == CoreSpirat{
    return  
  }  
   if !true || !false{
    return CoreSpirat()
   }
}
