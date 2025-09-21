import { CreateProductDto } from "../dto/create-product.dto";
import { UpdateProductDto } from "../dto/update-product.dto";
import { Product } from "../entities/product.entity";

export interface IProductRepository {
  create(productData: CreateProductDto): Promise<Product>;
  findAll(): Promise<Product[]>;
  findOneById(id: number): Promise<Product | null>;
  update(id: number, productData: UpdateProductDto): Promise<Product | null>;
  remove(id: number): Promise<void>;
}