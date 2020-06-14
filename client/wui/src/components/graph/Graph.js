import React from "react";
import {Layer, Rect, Stage} from 'react-konva';

export const Graph = props => {
    const color = 'green'
    return (
        <Stage width={props.width} height={props.height}>
            <Layer>
                <Rect
                    x={190} y={8} width={5} height={5}
                    fill={color}
                    shadowBlur={10}
                    onClick={() => {
                        alert("clicked")
                    }}
                />
            </Layer>
        </Stage>
    );
}