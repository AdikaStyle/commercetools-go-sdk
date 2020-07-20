package main

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
)

// Generate the `<service>DeleteWithID` and `<service>DeleteWithKey` functions
func deleteResourceHTTPMethod(resource *RamlType, resourceService ResourceService, resourceMethod ResourceMethod, httpMethod ResourceHTTPMethod) (code *jen.Statement) {
	methodName := fmt.Sprintf("%sDelete%s", resource.CodeName, strings.Title(resourceMethod.MethodName))
	resourceIdentifier := createResourceIdentifier(resourceService, resourceMethod)

	deleteWithVersion := true
	// TODO: nasty hack / incomplete API def
	if resourceService.ResourceType == "APIClient" {
		deleteWithVersion = false
	}

	methodParamList := []jen.Code{
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id(resourceIdentifier.ArgName).String(),
	}

	setVersionParam := jen.Empty()
	if deleteWithVersion {
		methodParamList = append(methodParamList, jen.Id("version").Int())
		setVersionParam = jen.Id("params").Op(".").Id("Set").Call(jen.Lit("version"), jen.Qual("strconv", "Itoa").Call(jen.Id("version")))
	}
	setDataErasure := jen.Empty()
	if httpMethod.HasTrait("dataErasure") {
		methodParamList = append(methodParamList, jen.Id("dataErasure").Bool())

		setDataErasure = jen.Id("params").Op(".").Id("Set").Call(jen.Lit("dataErasure"), jen.Qual("strconv", "FormatBool").Call(jen.Id("dataErasure")))
	}
	methodParamList = append(methodParamList, jen.Id("opts").Op("...").Id("RequestOption"))

	methodParams := jen.List(methodParamList...)
	clientMethod := "Delete"

	returnParams := jen.Id("client").Op("*").Id("Client")

	description := fmt.Sprintf("for type %s", resourceService.ResourceType)
	if httpMethod.Description != "" {
		description = httpMethod.Description
	}
	c := jen.Commentf("%s %s", methodName, description).Line()
	c.Func().Params(returnParams).Id(methodName).Params(methodParams).Parens(jen.List(jen.Id("result").Op("*").Id(resourceService.ResourceType), jen.Err().Error())).Block(
		jen.Id("params").Op(":=").Qual("net/url", "Values").Block(),
		setVersionParam,
		setDataErasure,

		// for _, opt := range opts {
		// 	opt(&params)
		// }
		jen.For(jen.List(jen.Id("_"), jen.Id("opt")).Op(":=").Range().Id("opts")).Block(
			jen.Id("opt").Call(jen.Op("&").Id("params")),
		),

		resourceIdentifier.createEndpointCode(false),
		jen.Id("err").Op("=").Id("client").Op(".").Id(clientMethod).Call(
			jen.Id("ctx"),
			jen.Id("endpoint"),
			jen.Id("params"),
			jen.Op("&").Id("result"),
		),
		jen.If(jen.Err().Op("!=").Nil()).Block(
			jen.Return(jen.Nil(), jen.Err()),
		),
		jen.Return(jen.Id("result"), jen.Nil()),
	).Line()

	return c
}