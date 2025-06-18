## API Response - Tipos Go lang

### Body - input: array
{
    model: string
    input: [
        {
            role: string
            type: string("message")
            content: 
            [ * ResponseInputContentUnionParam *
                { * ResponseInputTextParam *
                    type: string("input_text")
                    text: string                
                },                
            ]
            
        },        
    ]
}
